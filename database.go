package main

import (
	"context"
	"log"
	"net/url"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Datbase interface {
	Save()
	Init()
}

type DatabaseImpl struct {
	mongoClient    *mongo.Client
	url_queues_col *mongo.Collection
	url_meta_col   *mongo.Collection
}

func (dl *DatabaseImpl) Init() {
	uri := os.Getenv("DB_CONNECTION_STRING")

	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)

	if err != nil {
		panic(err)
	}

	dl.mongoClient = client
	dl.url_queues_col = client.Database("cedi_search").Collection("url_queues")
	dl.url_meta_col = client.Database("cedi_search").Collection("url_meta")

	// defer func() {
	// 	if err = client.Disconnect(context.TODO()); err != nil {
	// 		panic(err)
	// 	}
	// }()
	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		panic(err)
	}
	log.Println("Pinged your deployment. You successfully connected to MongoDB!")

}

func (dl *DatabaseImpl) SaveRawData(html string, u string) {
	os.WriteFile(url.PathEscape(u)+".html", []byte(html), 0644)
}

// QueueURL adds the given URL to the database's URL queue.
//
// url: a string representing the URL to be queued.
func (dl *DatabaseImpl) QueueURL(url string) {
	dl.url_queues_col.InsertOne(context.TODO(), url)
}

// GetURLMeta retrieves the URL meta data from the database.
//
// It takes a string parameter `url` which represents the URL to retrieve meta data for.
// The function does not return any value.
func (dl *DatabaseImpl) GetURLMeta(url string) {
	var meta interface{}
	coll := dl.url_meta_col
	filter := bson.D{{Key: "url", Value: url}}

	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &meta); err != nil {
		panic(err)
	}
	log.Println(meta)
}

// GetURLQueues retrieves URL queues from the database.
//
// No parameters.
// Returns a slice of URLQueue.
func (dl *DatabaseImpl) GetURLQueues() []URLQueue {
	var queues []URLQueue
	coll := dl.url_queues_col
	filter := bson.D{}

	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &queues); err != nil {
		panic(err)
	}

	return queues
}
