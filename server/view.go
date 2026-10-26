package main

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mattermost/mattermost/server/public/model"
)

// loadForMember returns a sent message if the user is in its channel.
func (p *Plugin) loadForMember(w http.ResponseWriter, r *http.Request) (*message, bool) {
	userID := r.Header.Get("Mattermost-User-Id")
	m, _, err := p.loadMessage(r.PathValue("id"))
	if err == nil && m.sent() {
		if _, appErr := p.API.GetChannelMember(m.ChannelID, userID); appErr == nil {
			return m, true
		}
	}
	http.Error(w, "not found", http.StatusNotFound)
	return nil, false
}

type viewerInfo struct {
	UserID   string `json:"user_id"`
	OpenedAt int64  `json:"opened_at"`
}

// handleInfo is what the post card shows. Only the sender learns who opened it.
func (p *Plugin) handleInfo(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	m, ok := p.loadForMember(w, r)
	if !ok {
		return
	}
	mine, opened := m.Views[userID]
	info := map[string]any{
		"is_sender":  userID == m.SenderID,
		"opened":     opened,
		"active":     opened && mine.active(model.GetMillis()),
		"has_text":   m.Text != "",
		"file_count": len(m.Files),
		"expires_at": m.ExpiresAt,
	}
	if userID == m.SenderID {
		recipients, err := p.recipients(m.ChannelID, m.SenderID)
		if err != nil {
			http.Error(w, "failed to list recipients", http.StatusInternalServerError)
			return
		}
		viewers := []viewerInfo{}
		for id, v := range m.Views {
			viewers = append(viewers, viewerInfo{UserID: id, OpenedAt: v.OpenedAt})
		}
		info["viewers"] = viewers
		info["recipient_count"] = len(recipients)
	}
	writeJSON(w, info)
}

type fileInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64  `json:"size"`
}

// handleOpen records the recipient's only view and returns the content. As in
// WhatsApp, the sender cannot see it again once sent.
func (p *Plugin) handleOpen(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	m, ok := p.loadForMember(w, r)
	if !ok {
		return
	}
	if userID == m.SenderID {
		http.Error(w, "the sender cannot open it", http.StatusForbidden)
		return
	}
	m, err := p.updateMessage(m.ID, func(m *message) error { return m.open(userID, model.GetMillis()) })
	if errors.Is(err, errAlreadyOpened) {
		http.Error(w, "already opened", http.StatusGone)
		return
	}
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	p.notifyChanged(m, userID)
	files := make([]fileInfo, 0, len(m.Files))
	for _, f := range m.Files {
		files = append(files, fileInfo{Name: f.Name, Type: f.Type, Size: f.Size})
	}
	writeJSON(w, map[string]any{"text": m.Text, "files": files})
}

// handleDownload serves a file while the user's viewer is open. It is always an
// attachment: the viewer fetches it and shows images and media from a blob.
func (p *Plugin) handleDownload(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	m, ok := p.loadForMember(w, r)
	if !ok {
		return
	}
	n, err := strconv.Atoi(r.PathValue("n"))
	if err != nil || n < 0 || n >= len(m.Files) || !m.canRead(userID, model.GetMillis()) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	f := m.Files[n]
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(f.Size, 10))
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(f.Name))
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if err := p.writeChunks(w, m.ID, n, f.Chunks); err != nil {
		p.API.LogError("Failed to serve a file", "id", m.ID, "err", err.Error())
	}
}

// handleClose ends the recipient's view; if they were the last one, the
// message goes away right now.
func (p *Plugin) handleClose(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	m, ok := p.loadForMember(w, r)
	if !ok {
		return
	}
	changed := false
	m, err := p.updateMessage(m.ID, func(m *message) error {
		changed = m.close(userID, model.GetMillis())
		return nil
	})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if changed {
		p.notifyChanged(m, userID)
		p.purgeIfFinished(m, model.GetMillis())
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDelete lets the sender delete the message for everybody, or drop a
// draft whose upload failed.
func (p *Plugin) handleDelete(w http.ResponseWriter, r *http.Request) {
	m, _, err := p.loadMessage(r.PathValue("id"))
	if err != nil || m.SenderID != r.Header.Get("Mattermost-User-Id") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	p.purge(m.ID, true)
	w.WriteHeader(http.StatusNoContent)
}
