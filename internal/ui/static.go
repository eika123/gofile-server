package ui

import (
	"embed"
	"net/http"
)

//go:embed css/* js/*
var staticFiles embed.FS

// StaticHandler returns an HTTP handler for embedded UI static assets.
func StaticHandler() http.Handler {
	return http.FileServer(http.FS(staticFiles))
}
