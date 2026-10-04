package web

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
)

//go:embed assets/*
var embedded embed.FS

// Handler отдаёт ресурсы и возвращает index.html для маршрутов приложения.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		isStatic := strings.HasPrefix(name, "static/")
		if isStatic {
			name = strings.TrimPrefix(name, "static/")
		}
		if name == "." || name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(embedded, "assets/"+name); err != nil {
			if isStatic {
				http.NotFound(w, r)
				return
			}
			if name == "api" || strings.HasPrefix(name, "api/") {
				http.NotFound(w, r)
				return
			}
			name = "index.html"
		}
		content, err := embedded.ReadFile("assets/" + name)
		if err != nil {
			slog.Error("failed to read embedded web asset", "asset", name, "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if name == "index.html" {
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		if strings.HasSuffix(name, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(name, ".js") {
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		if r.Method == http.MethodHead {
			return
		}
		if _, err := w.Write(content); err != nil {
			slog.Error("failed to write embedded web asset", "asset", name, "error", err)
		}
	})
}
