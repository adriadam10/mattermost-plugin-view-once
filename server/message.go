package main

import (
	"errors"
	"time"
)

// A recipient can read the content from opening it until closing the viewer.
// The window bounds that when the viewer is never closed (tab killed, crash).
const openWindow = 30 * time.Minute

var (
	errAlreadyOpened = errors.New("already opened")
	errNotSent       = errors.New("not sent yet")
)

// message is a view-once message: its text and the list of its files. The file
// contents are stored apart, in chunks, so the message itself stays small.
type message struct {
	ID        string          `json:"id"`
	SenderID  string          `json:"sender_id"`
	ChannelID string          `json:"channel_id"`
	RootID    string          `json:"root_id,omitempty"`
	Text      string          `json:"text"`
	Files     []fileMeta      `json:"files"`
	CreatedAt int64           `json:"created_at"`
	SentAt    int64           `json:"sent_at,omitempty"`
	PostID    string          `json:"post_id,omitempty"`
	ExpiresAt int64           `json:"expires_at"`
	Views     map[string]view `json:"views,omitempty"`
}

type fileMeta struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Chunks   int    `json:"chunks"`
	Complete bool   `json:"complete"`
}

type view struct {
	OpenedAt int64 `json:"opened_at"`
	ClosedAt int64 `json:"closed_at,omitempty"`
}

func (v view) active(now int64) bool {
	return v.ClosedAt == 0 && now < v.OpenedAt+openWindow.Milliseconds()
}

func (m *message) sent() bool {
	return m.SentAt != 0
}

// canRead: only a recipient while their viewer is open. The sender never
// opens it, so has no view.
func (m *message) canRead(userID string, now int64) bool {
	v, ok := m.Views[userID]
	return ok && v.active(now)
}

// open records the one and only view of a recipient.
func (m *message) open(userID string, now int64) error {
	if !m.sent() {
		return errNotSent
	}
	if _, ok := m.Views[userID]; ok {
		return errAlreadyOpened
	}
	if m.Views == nil {
		m.Views = map[string]view{}
	}
	m.Views[userID] = view{OpenedAt: now}
	return nil
}

// close ends a recipient's view; it reports whether anything changed.
func (m *message) close(userID string, now int64) bool {
	v, ok := m.Views[userID]
	if !ok || v.ClosedAt != 0 {
		return false
	}
	v.ClosedAt = now
	m.Views[userID] = v
	return true
}

// finished: every recipient has opened it and nobody is still looking at it.
func (m *message) finished(recipients []string, now int64) bool {
	if len(recipients) == 0 {
		return false
	}
	for _, userID := range recipients {
		v, ok := m.Views[userID]
		if !ok || v.active(now) {
			return false
		}
	}
	return true
}

func (m *message) uploadsComplete() bool {
	for _, f := range m.Files {
		if !f.Complete {
			return false
		}
	}
	return true
}
