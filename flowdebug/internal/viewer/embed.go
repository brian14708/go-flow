package viewer

import (
	"embed"
	"net/http"
)

//go:embed profiler/**
var assets embed.FS

func Handler() http.Handler {
	return http.FileServer(http.FS(assets))
}
