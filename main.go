package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path"
)

//go:embed public
var files embed.FS

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fsys, err := fs.Sub(files, "public")
		if err != nil {
			log.Fatal(err)
		}
		p := r.URL.Path
		if path.Ext(p) == "" {
			p = p + ".html"
			w.Header().Set("Content-Type", "text/html")
		}
		if path.Ext(p) == ".css" {
			w.Header().Set("Content-Type", "text/css")
		}
		content, err := fs.ReadFile(fsys, p[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(content)
	})

	log.Fatal(http.ListenAndServe(":8080", mux))
}
