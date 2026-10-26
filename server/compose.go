package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"time"
	"unicode/utf8"

	"github.com/mattermost/mattermost/server/public/model"
)

type createRequest struct {
	ChannelID string `json:"channel_id"`
	RootID    string `json:"root_id"`
	Text      string `json:"text"`
}

// handleCreate starts a draft with the text; the files follow one per request
// and handleSend posts it.
func (p *Plugin) handleCreate(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	var req createRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256*1024)).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if utf8.RuneCountInString(req.Text) > model.PostMessageMaxRunesV2 {
		http.Error(w, "text too long", http.StatusRequestEntityTooLarge)
		return
	}
	if err := p.checkCanPost(userID, req.ChannelID, req.RootID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	now := model.GetMillis()
	m := &message{
		ID:        model.NewId(),
		SenderID:  userID,
		ChannelID: req.ChannelID,
		RootID:    req.RootID,
		Text:      req.Text,
		Files:     []fileMeta{},
		CreatedAt: now,
		ExpiresAt: now + draftLifetime.Milliseconds(),
	}
	if err := p.createMessage(m); err != nil {
		p.API.LogError("Failed to create a view-once message", "err", err.Error())
		http.Error(w, "failed to save", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"id": m.ID})
}

func (p *Plugin) checkCanPost(userID, channelID, rootID string) error {
	channel, appErr := p.API.GetChannel(channelID)
	if appErr != nil || channel.DeleteAt != 0 {
		return errors.New("channel not found")
	}
	if !p.API.HasPermissionToChannel(userID, channelID, model.PermissionCreatePost) {
		return errors.New("you cannot post in this channel")
	}
	if rootID == "" {
		return nil
	}
	root, appErr := p.API.GetPost(rootID)
	if appErr != nil || root.ChannelId != channelID || root.RootId != "" {
		return errors.New("thread not found")
	}
	return nil
}

// loadOwnDraft returns a message of the user that is not sent yet.
func (p *Plugin) loadOwnDraft(w http.ResponseWriter, r *http.Request) (*message, bool) {
	m, _, err := p.loadMessage(r.PathValue("id"))
	if err != nil || m.SenderID != r.Header.Get("Mattermost-User-Id") {
		http.Error(w, "not found", http.StatusNotFound)
		return nil, false
	}
	if m.sent() {
		http.Error(w, "already sent", http.StatusConflict)
		return nil, false
	}
	return m, true
}

func (p *Plugin) maxFileSize() int64 {
	if size := p.API.GetConfig().FileSettings.MaxFileSize; size != nil {
		return *size
	}
	return defaultMaxFileSize
}

// handleUpload adds one file, sent as the raw request body, with its name and
// type in the query string.
func (p *Plugin) handleUpload(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	m, ok := p.loadOwnDraft(w, r)
	if !ok {
		return
	}
	if !p.API.HasPermissionToChannel(userID, m.ChannelID, model.PermissionUploadFile) {
		http.Error(w, "you cannot upload files in this channel", http.StatusForbidden)
		return
	}
	name := cleanName(r.URL.Query().Get("name"))
	contentType := r.URL.Query().Get("type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Reserve the file's slot first, so parallel uploads do not share chunk keys.
	var index int
	m, err := p.updateMessage(m.ID, func(m *message) error {
		if m.sent() {
			return errors.New("already sent")
		}
		if len(m.Files) >= maxFiles {
			return fmt.Errorf("at most %d files", maxFiles)
		}
		index = len(m.Files)
		m.Files = append(m.Files, fileMeta{Name: name, Type: contentType})
		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	size, chunks, err := p.storeChunks(m, index, r.Body, p.maxFileSize())
	if errors.Is(err, errTooLarge) {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}
	if err != nil {
		p.API.LogError("Failed to store a file", "id", m.ID, "err", err.Error())
		http.Error(w, "failed to save the file", http.StatusInternalServerError)
		return
	}
	_, err = p.updateMessage(m.ID, func(m *message) error {
		if m.sent() {
			return errors.New("already sent")
		}
		m.Files[index].Size = size
		m.Files[index].Chunks = chunks
		m.Files[index].Complete = true
		return nil
	})
	if err != nil {
		p.deleteChunks(m.ID, index, chunks)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func cleanName(name string) string {
	name = path.Base("/" + name)
	if name == "/" || name == "." {
		name = "file"
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		name = string([]rune(name)[:maxNameLength])
	}
	return name
}

// handleSend posts the placeholder in the channel and starts the countdown.
func (p *Plugin) handleSend(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	m, ok := p.loadOwnDraft(w, r)
	if !ok {
		return
	}
	if m.Text == "" && len(m.Files) == 0 {
		http.Error(w, "empty message", http.StatusBadRequest)
		return
	}
	if err := p.checkCanPost(userID, m.ChannelID, m.RootID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	now := model.GetMillis()
	lifetime := time.Duration(p.settings.maxLifetimeDays.Load()) * 24 * time.Hour
	m, err := p.updateMessage(m.ID, func(m *message) error {
		if m.sent() {
			return errors.New("already sent")
		}
		if !m.uploadsComplete() {
			return errors.New("a file is still uploading")
		}
		m.SentAt = now
		m.ExpiresAt = now + lifetime.Milliseconds()
		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	locale := ""
	if sender, appErr := p.API.GetUser(userID); appErr == nil {
		locale = sender.Locale
	}
	post, appErr := p.API.CreatePost(&model.Post{
		UserId:    userID,
		ChannelId: m.ChannelID,
		RootId:    m.RootID,
		Type:      postType,
		Message:   placeholder(locale),
		Props:     model.StringInterface{propID: m.ID},
	})
	if appErr != nil {
		p.API.LogError("Failed to post a view-once message", "id", m.ID, "err", appErr.Error())
		p.purge(m.ID, true)
		http.Error(w, "failed to post", http.StatusInternalServerError)
		return
	}
	if _, err := p.updateMessage(m.ID, func(m *message) error {
		m.PostID = post.Id
		return nil
	}); err != nil {
		p.API.LogError("Failed to link a view-once message to its post", "id", m.ID, "err", err.Error())
	}
	writeJSON(w, map[string]string{"post_id": post.Id})
}
