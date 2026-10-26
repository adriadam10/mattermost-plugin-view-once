package main

import (
	"errors"
	"testing"
)

func sentMessage() *message {
	return &message{SenderID: "sender", SentAt: 1}
}

func TestOpenOnlyOnce(t *testing.T) {
	m := sentMessage()
	if err := m.open("bob", 100); err != nil {
		t.Fatalf("first open: %v", err)
	}
	if err := m.open("bob", 200); !errors.Is(err, errAlreadyOpened) {
		t.Fatalf("second open: got %v, want errAlreadyOpened", err)
	}
}

func TestOpenDraft(t *testing.T) {
	m := &message{SenderID: "sender"}
	if err := m.open("bob", 100); !errors.Is(err, errNotSent) {
		t.Fatalf("got %v, want errNotSent", err)
	}
}

func TestCanRead(t *testing.T) {
	m := sentMessage()
	window := openWindow.Milliseconds()
	if m.canRead("sender", 0) {
		t.Error("the sender cannot read it once sent")
	}
	if m.canRead("bob", 100) {
		t.Error("not opened yet")
	}
	_ = m.open("bob", 100)
	if !m.canRead("bob", 100+window-1) {
		t.Error("inside the window")
	}
	if m.canRead("bob", 100+window) {
		t.Error("the window is over")
	}
	m.close("bob", 150)
	if m.canRead("bob", 151) {
		t.Error("closed")
	}
}

func TestFinished(t *testing.T) {
	m := sentMessage()
	recipients := []string{"bob", "ann"}
	if m.finished(nil, 0) {
		t.Error("nobody to read it is not finished")
	}
	_ = m.open("bob", 100)
	m.close("bob", 110)
	if m.finished(recipients, 120) {
		t.Error("ann has not opened it")
	}
	_ = m.open("ann", 130)
	if m.finished(recipients, 140) {
		t.Error("ann is still looking at it")
	}
	if !m.finished(recipients, 130+openWindow.Milliseconds()) {
		t.Error("ann's window is over")
	}
	m.close("ann", 150)
	if !m.finished(recipients, 160) {
		t.Error("everybody opened and closed it")
	}
}
