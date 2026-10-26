package main

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	// The post in the channel only says there is a view-once message; the
	// content never reaches posts, file attachments, notifications or search.
	postType  = "custom_view_once"
	propID    = "view_once_id"
	sweepTick = 5 * time.Minute

	// Sent to the sender and the viewer when a message is opened, closed or deleted.
	changedEvent = "changed"
)

type Plugin struct {
	plugin.MattermostPlugin
	router   http.Handler
	settings settings
	stop     chan struct{}
	done     sync.WaitGroup
}

func (p *Plugin) OnActivate() error {
	p.router = p.newRouter()
	p.stop = make(chan struct{})
	p.done.Go(p.sweepLoop)
	return nil
}

func (p *Plugin) OnDeactivate() error {
	close(p.stop)
	p.done.Wait()
	return nil
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

// MessageWillBePosted only lets through view-once posts that the plugin itself
// is creating, so nobody can forge one through the API.
func (p *Plugin) MessageWillBePosted(_ *plugin.Context, post *model.Post) (*model.Post, string) {
	if post.Type != postType {
		return post, ""
	}
	id, _ := post.GetProp(propID).(string)
	m, _, err := p.loadMessage(id)
	if err != nil || !m.sent() || m.PostID != "" || m.SenderID != post.UserId || m.ChannelID != post.ChannelId {
		return nil, "View-once messages can only be sent with the view-once button."
	}
	return post, ""
}

// MessageWillBeUpdated keeps the post as it was: there is nothing to edit in
// it, but pinning or other changes must still work.
func (p *Plugin) MessageWillBeUpdated(_ *plugin.Context, newPost, oldPost *model.Post) (*model.Post, string) {
	if oldPost.Type != postType {
		return newPost, ""
	}
	newPost.Type = oldPost.Type
	newPost.Message = oldPost.Message
	newPost.AddProp(propID, oldPost.GetProp(propID))
	return newPost, ""
}

// MessageHasBeenDeleted removes the content when the sender (or an admin)
// deletes the post from its menu.
func (p *Plugin) MessageHasBeenDeleted(_ *plugin.Context, post *model.Post) {
	if post.Type != postType {
		return
	}
	if id, ok := post.GetProp(propID).(string); ok {
		p.purge(id, false)
	}
}

func (p *Plugin) sweepLoop() {
	ticker := time.NewTicker(sweepTick)
	defer ticker.Stop()
	for {
		p.sweep()
		select {
		case <-p.stop:
			return
		case <-ticker.C:
		}
	}
}

// sweep deletes expired messages, unsent drafts and messages that every
// recipient has already seen (the last viewer may never have closed it).
func (p *Plugin) sweep() {
	ids, err := p.messageIDs()
	if err != nil {
		p.API.LogError("Failed to list view-once messages", "err", err.Error())
		return
	}
	now := model.GetMillis()
	for _, id := range ids {
		m, _, err := p.loadMessage(id)
		if err != nil {
			continue
		}
		if now >= m.ExpiresAt {
			p.purge(id, true)
			continue
		}
		if len(m.Views) > 0 {
			p.purgeIfFinished(m, now)
		}
	}
}

func (p *Plugin) purgeIfFinished(m *message, now int64) {
	recipients, err := p.recipients(m.ChannelID, m.SenderID)
	if err != nil {
		p.API.LogWarn("Failed to list recipients", "id", m.ID, "err", err.Error())
		return
	}
	if m.finished(recipients, now) {
		p.purge(m.ID, true)
	}
}

// recipients are the channel members who can open the message: everybody but
// the sender, bots and deactivated users (they stay members of their channels).
func (p *Plugin) recipients(channelID, senderID string) ([]string, error) {
	const perPage = 200
	var ids []string
	for page := 0; ; page++ {
		users, appErr := p.API.GetUsersInChannel(channelID, model.ChannelSortByUsername, page, perPage)
		if appErr != nil {
			return nil, appErr
		}
		for _, u := range users {
			if u.Id != senderID && !u.IsBot && u.DeleteAt == 0 {
				ids = append(ids, u.Id)
			}
		}
		if len(users) < perPage {
			return ids, nil
		}
	}
}

func (p *Plugin) notifyChanged(m *message, viewerID string) {
	for _, userID := range []string{m.SenderID, viewerID} {
		if userID == "" {
			continue
		}
		p.API.PublishWebSocketEvent(changedEvent, map[string]any{"id": m.ID}, &model.WebsocketBroadcast{UserId: userID})
	}
}

// placeholder is the text of the post, which is all that mobile apps and
// notifications show.
func placeholder(locale string) string {
	if strings.HasPrefix(locale, "es") {
		return "🔒 Mensaje para ver una vez. Ábrelo desde Mattermost en el navegador o en el escritorio."
	}
	return "🔒 View-once message. Open it from Mattermost on the web or the desktop app."
}

func main() {
	plugin.ClientMain(&Plugin{})
}
