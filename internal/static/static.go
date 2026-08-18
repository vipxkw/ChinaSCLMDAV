package static

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

//go:embed dist/*
var distFS embed.FS

// Handler serves the built frontend (single-page app).
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upath := strings.TrimPrefix(r.URL.Path, "/")
		if upath == "" {
			upath = "index.html"
		}
		p := path.Clean("/" + upath)
		data, err := fs.ReadFile(sub, strings.TrimPrefix(p, "/"))
		if err != nil {
			// SPA fallback: unknown non-asset routes get index.html.
			if !strings.HasPrefix(p, "/assets/") {
				p = "/index.html"
				data, err = fs.ReadFile(sub, "index.html")
			}
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		if ct := mime.TypeByExtension(path.Ext(p)); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		// Assets are embedded at build time; always revalidate to avoid stale UI.
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(data)
	})
}