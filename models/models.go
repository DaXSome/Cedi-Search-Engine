package models

type UrlQueue struct {
	URL    string `json:"url"`
	Source string `json:"source"`
}

type CrawledPage struct {
	URL    string `json:"url"`
	HTML   string `json:"html"`
	Source string `json:"source"`
}
