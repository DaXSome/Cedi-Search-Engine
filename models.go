package main

type URL struct {
	URLString string  `bson:"url_string"`
	ItemTag   ItemTag `bson:"item_tag"`
}

type ItemTag struct {
	Attr        string `bson:"attr"`
	ValuePrefix string `bson:"value_prefix"`
}

type URLQueue struct {
	ID      string `bson:"_id"`
	URLItem URL    `bson:"url_item"`
}

type RawData struct {
	ID   string `bson:"_id"`
	HTML string `bson:"html"`
}
