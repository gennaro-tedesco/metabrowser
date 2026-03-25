package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

type View struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	BookPaths []string `json:"bookPaths"`
}

func viewsPath(root string) string {
	sum := sha256.Sum256([]byte(root))
	return filepath.Join(cacheDir(), hex.EncodeToString(sum[:])+"-views.json")
}

func loadViews(root string) []View {
	if root == "" {
		return []View{}
	}
	data, err := os.ReadFile(viewsPath(root))
	if err != nil {
		return []View{}
	}
	var views []View
	if err := json.Unmarshal(data, &views); err != nil {
		return []View{}
	}
	return views
}

func saveViews(root string, views []View) error {
	if root == "" {
		return nil
	}
	p := viewsPath(root)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(views, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}
