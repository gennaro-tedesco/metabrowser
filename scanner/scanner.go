package scanner

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"calibre-browser/model"
)

func Scan(root string) (model.Library, error) {
	var epubs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.ToLower(filepath.Ext(path)) == ".epub" {
			epubs = append(epubs, path)
		}
		return nil
	})
	if err != nil {
		return model.Library{}, err
	}

	numWorkers := runtime.NumCPU()
	jobs := make(chan string, len(epubs))
	results := make(chan model.Book, len(epubs))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				book := parseBook(root, path)
				results <- book
			}
		}()
	}

	for _, e := range epubs {
		jobs <- e
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var books []model.Book
	for b := range results {
		books = append(books, b)
	}

	sort.Slice(books, func(i, j int) bool {
		return books[i].Title < books[j].Title
	})

	return buildLibrary(books), nil
}

func parseBook(root, path string) model.Book {
	title, language, series, coverPath, authors, tags, err := parseEpub(path)
	if err != nil || title == "" {
		title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	rel, _ := filepath.Rel(root, path)

	coverURL := ""
	if coverPath != "" {
		coverURL = "/cover/" + filepath.ToSlash(rel) + "?entry=" + coverPath
	}

	return model.Book{
		Title:     title,
		Authors:   authors,
		Language:  normaliseLanguage(language),
		Series:    series,
		Tags:      tags,
		Path:      rel,
		CoverURL:  coverURL,
		CoverPath: coverPath,
	}
}

func buildLibrary(books []model.Book) model.Library {
	langSet := map[string]struct{}{}
	seriesSet := map[string]struct{}{}
	tagSet := map[string]struct{}{}
	authorSet := map[string]struct{}{}

	for _, b := range books {
		if b.Language != "" {
			langSet[b.Language] = struct{}{}
		}
		if b.Series != "" {
			seriesSet[b.Series] = struct{}{}
		}
		for _, t := range b.Tags {
			tagSet[t] = struct{}{}
		}
		for _, a := range b.Authors {
			authorSet[a] = struct{}{}
		}
	}

	groups := map[string][]string{
		"language": setToSortedSlice(langSet),
		"series":   setToSortedSlice(seriesSet),
		"tags":     setToSortedSlice(tagSet),
		"author":   setToSortedSlice(authorSet),
	}

	return model.Library{Books: books, Groups: groups}
}

func setToSortedSlice(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func normaliseLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch lang {
	case "en", "eng":
		return "English"
	case "de", "deu", "ger":
		return "German"
	case "fr", "fra", "fre":
		return "French"
	case "it", "ita":
		return "Italian"
	case "es", "spa":
		return "Spanish"
	case "pt", "por":
		return "Portuguese"
	case "nl", "nld", "dut":
		return "Dutch"
	case "ru", "rus":
		return "Russian"
	case "zh", "zho", "chi":
		return "Chinese"
	case "ja", "jpn":
		return "Japanese"
	case "ar", "ara":
		return "Arabic"
	default:
		if lang == "" {
			return "Unknown"
		}
		return strings.Title(lang)
	}
}
