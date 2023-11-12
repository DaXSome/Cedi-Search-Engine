package main

type UrlQueue struct {
	URL string `json:"url"`
}

type CrawledPage struct {
	URL    string `json:"url"`
	HTML   string `json:"html"`
	Source string `json:"source"`
}
