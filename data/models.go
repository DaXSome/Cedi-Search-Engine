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
	ProductID   string   `bson:"_id" json:"objectID"`
	Slug        string   `bson:"slug" json:"slug"`
	Name        string   `bson:"name" json:"name"`
	Price       float64  `bson:"price" json:"price"`
	Rating      float64  `bson:"rating" json:"rating"`
	Description string   `bson:"description" json:"description"`
	URL         string   `bson:"url" json:"url"`
	Source      string   `bson:"source" json:"source"`
	Images      []string `bson:"images" json:"images"`
}

type AlgoliaData struct {
	ObjectID string `json:"objectID"`
	Product
}
