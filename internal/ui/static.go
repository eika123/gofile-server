package ui

import (
	"embed"
	"net/http"
)

//go:embed css/* js/*
var staticFiles embed.FS

// StaticHandler returns an HTTP handler for embedded UI static assets.
func StaticHandler() http.Handler {
	fs := http.FileServer(http.FS(staticFiles))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		fs.ServeHTTP(w, r)
	})
}
