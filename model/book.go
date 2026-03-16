package model

type Book struct {
	Title    string   `json:"title"`
	Authors  []string `json:"authors"`
	Language string   `json:"language"`
	Series   string   `json:"series"`
	Tags     []string `json:"tags"`
	Path     string   `json:"path"`
	CoverURL string   `json:"coverUrl"`
}

type Library struct {
	Books  []Book              `json:"books"`
	Groups map[string][]string `json:"groups"`
}
