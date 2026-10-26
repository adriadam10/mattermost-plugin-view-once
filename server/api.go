package main

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	// The same as a regular post.
	maxFiles = 10
	// A draft that is never sent (the browser closed while uploading) goes away soon.
	draftLifetime = time.Hour
	maxNameLength = 255
	// Mattermost's own default, in case the setting is missing.
	defaultMaxFileSize = 100 << 20
)

func (p *Plugin) newRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/messages", p.handleCreate)
	mux.HandleFunc("POST /api/v1/messages/{id}/files", p.handleUpload)
	mux.HandleFunc("POST /api/v1/messages/{id}/send", p.handleSend)
	mux.HandleFunc("GET /api/v1/messages/{id}", p.handleInfo)
	mux.HandleFunc("POST /api/v1/messages/{id}/open", p.handleOpen)
	mux.HandleFunc("GET /api/v1/messages/{id}/files/{n}", p.handleDownload)
	mux.HandleFunc("POST /api/v1/messages/{id}/close", p.handleClose)
	mux.HandleFunc("DELETE /api/v1/messages/{id}", p.handleDelete)
	return requireUser(mux)
}

func requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Mattermost-User-Id") == "" {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}
