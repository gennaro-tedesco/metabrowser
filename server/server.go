package server

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"calibre-browser/model"
)

func Serve(addr, root string, lib model.Library, static fs.FS) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/library", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(lib)
	})

	mux.HandleFunc("/cover/", func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/cover/")
		abs := filepath.Join(root, filepath.FromSlash(rel), "cover.jpg")
		if _, err := os.Stat(abs); err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, abs)
	})

	stripped, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(stripped)))

	log.Printf("calibre-browser listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
