package data

type UrlQueue struct {
	ID     string `bson:"_id"`
	URL    string `bson:"url"`
	Source string `bson:"source"`
}

type CrawledPage struct {
	URL    string `bson:"url"`
	HTML   string `bson:"html"`
	Source string `bson:"source"`
}

type Product struct {
	ProductID   string   `bson:"_id"`
	Name        string   `bson:"name"`
	Price       float64  `bson:"price"`
	Rating      float64  `bson:"rating"`
	Description string   `bson:"description"`
	URL         string   `bson:"url"`
	Source      string   `bson:"source"`
	Images      []string `bson:"images"`
}

type AlgoliaData struct {
	ObjectID string `json:"objectID"`
	Product
}
