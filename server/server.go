package server

import (
	"archive/zip"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"calibre-browser/model"
)

func Serve(addr, root string, lib model.Library, static fs.FS) {
	epubIndex := make(map[string]string, len(lib.Books))
	for _, b := range lib.Books {
		epubIndex[b.Path] = b.CoverPath
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/library", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(lib)
	})

	mux.HandleFunc("/cover/", func(w http.ResponseWriter, r *http.Request) {
		epubRel := strings.TrimPrefix(r.URL.Path, "/cover/")
		entry := r.URL.Query().Get("entry")
		if entry == "" {
			http.NotFound(w, r)
			return
		}

		absEpub := filepath.Join(root, filepath.FromSlash(epubRel))
		zr, err := zip.OpenReader(absEpub)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer zr.Close()

		for _, f := range zr.File {
			if f.Name != entry {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				http.Error(w, "error reading cover", http.StatusInternalServerError)
				return
			}
			defer rc.Close()
			w.Header().Set("Content-Type", mediaTypeForEntry(entry))
			w.Header().Set("Cache-Control", "public, max-age=86400")
			io.Copy(w, rc)
			return
		}
		http.NotFound(w, r)
	})

	stripped, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(stripped)))

	log.Printf("metabrowser listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func mediaTypeForEntry(entry string) string {
	switch strings.ToLower(filepath.Ext(entry)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}
