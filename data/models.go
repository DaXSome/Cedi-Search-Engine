package data

type UrlQueue struct {
	URL    string `json:"url"`
	Source string `json:"source"`
}

type CrawledPage struct {
	URL    string `json:"url"`
	HTML   string `json:"html"`
	Source string `json:"source"`
}

type Product struct {
	ProductID   string   `json:"product_id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Rating      float64  `json:"rating"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Source      string   `json:"source"`
	Images      []string `json:"images"`
}

type AlgoliaData struct {
	ObjectID string `json:"objectID"`
	Product
}
