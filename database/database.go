package database

import (
	"context"
	"fmt"
	"math/rand"
	netURL "net/url"
	"os"
	"strings"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	*mongo.Database
	AlgoliaIndex *search.Index
}

// NewDatabase initializes a new instance of the Database struct.
//
// Returns a pointer to the newly created Database.
func NewDatabase() *Database {
	utils.Logger(utils.Database, utils.Database, "Initing database...")

	dbURI := os.Getenv("DB_URI")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dbURI))
	if err != nil {
		utils.Logger(utils.Error, utils.Database, err)
	}

	algoliaClient := search.NewClient(os.Getenv("ALGOLIA_APP_ID"), os.Getenv("ALGOLIA_API_KEY"))

	algoliaIndex := algoliaClient.InitIndex("products")

	utils.Logger(utils.Database, utils.Database, "Database initialized!")

	return &Database{
		AlgoliaIndex: algoliaIndex,
		Database:     client.Database("cedi_search"),
	}
}

// GetQueue retrieves a slice of data.UrlQueue from the Database.
// It randomly selects 10 URLs from the queue and returns
// them as a slice of data.UrlQueue.
func (db *Database) GetQueue(source string) ([]data.UrlQueue, error) {
	utils.Logger(utils.Database, utils.Database, "Getting queue for ", source)

	queueCol := db.Collection("url_queues")

	queueCount, err := queueCol.CountDocuments(context.TODO(), bson.D{{Key: "source", Value: source}}, options.Count())
	if err != nil {
		return []data.UrlQueue{}, err
	}

	if queueCount == 0 {
		return []data.UrlQueue{}, nil
	}

	skipN := rand.Intn(int(queueCount))

	res, err := queueCol.Find(
		context.TODO(),
		bson.D{{Key: "source", Value: source}},
		&options.FindOptions{
			Skip:  options.Count().SetSkip(int64(skipN)).Skip,
			Limit: options.Count().SetLimit(5).Limit,
		})
	if err != nil {
		return []data.UrlQueue{}, err
	}

	var queues []data.UrlQueue
	res.All(context.TODO(), &queues)

	return queues, nil
}

// AddToQueue adds a URL to the queue in the Database.
//
// It takes a parameter 'url' of type `data.UrlQueue` which represents the URL to be added.
func (db *Database) AddToQueue(url data.UrlQueue) error {
	utils.Logger(utils.Database, utils.Database, "Adding to queue...", url.URL)

	parsedURL, err := netURL.Parse(url.URL)
	if err != nil {
		return err
	}

	url.ID = parsedURL.Path

	_, err = db.Collection("url_queues").InsertOne(context.TODO(), url, &options.InsertOneOptions{})
	if err != nil {
		return err
	}

	utils.Logger(utils.Database, utils.Database, "Added to queue!")

	return nil
}

// DeleteFromQueue deletes a URL from the queue in the Database.
//
// It takes a parameter `url` of type `data.UrlQueue`, which represents the URL to be deleted from the queue.
// This function does not return any value.
func (db *Database) DeleteFromQueue(url data.UrlQueue) error {
	utils.Logger(utils.Database, utils.Database, "Deleting from queue...", url.URL)

	_, err := db.Collection("url_queues").DeleteOne(context.TODO(), url)
	if err != nil {
		return err
	}

	utils.Logger(utils.Database, utils.Database, "Deleted from queue")

	return nil
}

// CanQueueUrl checks if a URL can be queued.
//
// Parameters:
// - url: the URL to check.
//
// Returns:
// - bool: true if the URL can be queued, false otherwise.
func (db *Database) CanQueueUrl(url string) (bool, error) {
	parsedURL, err := netURL.Parse(url)
	if err != nil {
		return false, err
	}

	existsInQueue := db.Collection("url_queues").FindOne(context.TODO(), bson.D{{Key: "_id", Value: parsedURL.Path}}).Err() == nil
	existsInIndexedProducts := db.Collection("indexed_products").FindOne(context.TODO(), bson.D{{Key: "_id", Value: parsedURL.Path}}) == nil

	canQueue := !existsInQueue && !existsInIndexedProducts

	utils.Logger(utils.Database, utils.Database, fmt.Sprintf("Can queue %s?", url), canQueue)

	return canQueue, nil
}

// GetCrawledPages retrieves crawled pages for a given source.
//
// Parameters:
// - source: a string representing the source of the crawled pages. e.g. Jumia
//
// Returns:
// - an array of data.CrawledPage representing the retrieved crawled pages.
func (db *Database) GetCrawledPages(source string) ([]data.CrawledPage, error) {
	utils.Logger(utils.Database, utils.Database, "Getting crawled pages for ", source)

	res, err := db.Collection("crawled_pages").Find(context.TODO(), bson.D{{Key: "source", Value: source}}, &options.FindOptions{Limit: options.Count().SetLimit(5).Limit})
	if err != nil {
		return []data.CrawledPage{}, err
	}

	var pages []data.CrawledPage
	res.All(context.TODO(), &pages)

	utils.Logger(utils.Database, utils.Database, "Crawled pages for ", source, " retrieved!")

	return pages, nil
}

// IndexProduct saves a product to the indexed_products collection in the database.
//
// It takes a parameter `product` of type `data.Product`.
func (db *Database) IndexProduct(product map[string]interface{}) error {
	utils.Logger(utils.Database, utils.Database, "Saving product...", product["name"])

	parsedURL, err := netURL.Parse(product["url"].(string))
	if err != nil {
		return err
	}

	segments := strings.Split(parsedURL.Path, "/")

	product["slug"] = segments[len(segments)-1]

	_, err = db.Collection("indexed_products").InsertOne(context.TODO(), product, &options.InsertOneOptions{})
	if err != nil {
		return err
	}

	res, err := db.AlgoliaIndex.SaveObject(product)
	if err != nil {
		return err
	}

	res.Wait()

	utils.Logger(utils.Database, utils.Database, "Product Saved!")

	_, err = db.Collection("meta_data").UpdateOne(context.TODO(), bson.M{"_id": "updated_at"}, bson.M{"$set": data.MetaData{UpdatedAt: time.Now().Format(time.RFC3339)}})
	if err != nil {
		return err
	}

	return nil
}
