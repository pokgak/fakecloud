// Package web serves the embedded visualizer dashboard.
package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var content embed.FS

func Handler() http.Handler {
	static, err := fs.Sub(content, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(static))
}
