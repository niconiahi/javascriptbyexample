package main

import (
	"embed"
	"fmt"
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
		fmt.Printf("%v\n", p)
		if path.Ext(p) == "" {
			fmt.Printf("is html %v\n", p)
			p = p + ".html"
			w.Header().Set("Content-Type", "text/html")
		}
		if path.Ext(p) == ".css" {
			fmt.Printf("is css %v\n", p)
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
