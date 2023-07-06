package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type URL struct {
	URL     string  `bson:"url"`
	ItemTag ItemTag `bson:"item_tag"`
}

type ItemTag struct {
	Attr        string `bson:"attr"`
	ValuePrefix string `bson:"value_prefix"`
}

type URLQueue struct {
	ID  primitive.ObjectID `bson:"_id"`
	URL URL                `bson:"url"`
}
