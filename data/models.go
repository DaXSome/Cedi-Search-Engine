package data

type UrlQueue struct {
	ID     string `bson:"_id"`
	URL    string `bson:"url"`
	Source string `bson:"source"`
}

type CrawledPage struct {
	URL     string `bson:"url"`
	HTML    string `bson:"html"`
	Source  string `bson:"source"`
	Attribs []Data
}

type Product struct {
	Slug        string   `bson:"slug" json:"slug"`
	Name        string   `bson:"name" json:"name"`
	Price       float64  `bson:"price" json:"price"`
	Rating      float64  `bson:"rating" json:"rating"`
	Description string   `bson:"description" json:"description"`
	URL         string   `bson:"url" json:"url"`
	Source      string   `bson:"source" json:"source"`
	Images      []string `bson:"images" json:"images"`
}

type MetaData struct {
	UpdatedAt string `bson:"updated_at"`
}

type AlgoliaData struct {
	ObjectID string `json:"objectID"`
	Product
}

type Selector struct {
	Element   string `json:"element"`
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

type Data struct {
	Label       string   `json:"label"`
	IsArray     bool     `json:"isArray"`
	ChildAttrib string   `json:"childAttrib"`
	Selector    Selector `json:"selector"`
}

type Target struct {
	Target   string `json:"target"`
	Host     string `json:"host"`
	SeedPath string `json:"seed_path"`
	Data     []Data `json:"data"`
}

type Config struct {
	Targets []Target `json:"targets"`
}
