package database

import (
	"context"
	"log"
	netURL "net/url"
	"os"
	"strings"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	*mongo.Database
}

// NewDatabase initializes a new instance of the Database struct.
//
// Returns a pointer to the newly created Database.
func NewDatabase() *Database {
	log.Println("[+] Initing database...")

	dbURI := os.Getenv("DB_URI")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dbURI))
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[+] Database initialized!")

	return &Database{
		client.Database("cedi_search"),
	}
}

// GetQueue retrieves a slice of data.UrlQueue from the Database.
// It randomly selects 10 URLs from the queue and returns
// them as a slice of data.UrlQueue.
func (db *Database) GetQueue(source string) []data.UrlQueue {
	log.Printf("[+] Getting queue for %s\n", source)

	res, err := db.Collection("url_queues").Find(context.TODO(), bson.D{{Key: "source", Value: source}}, &options.FindOptions{Limit: options.Count().SetLimit(5).Limit})
	if err != nil {
		log.Fatalln(err)
	}

	var queues []data.UrlQueue
	res.All(context.TODO(), &queues)

	return queues
}

// AddToQueue adds a URL to the queue in the Database.
//
// It takes a parameter 'url' of type `data.UrlQueue` which represents the URL to be added.
func (db *Database) AddToQueue(url data.UrlQueue) {
	log.Println("[+] Adding to queue...", url.URL)

	parsedURL, err := netURL.Parse(url.URL)
	if err != nil {
		log.Fatalln(err)
	}

	url.ID = parsedURL.Path

	_, err = db.Collection("url_queues").InsertOne(context.TODO(), url, &options.InsertOneOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[+] Added to queue!")
}

// DeleteFromQueue deletes a URL from the queue in the Database.
//
// It takes a parameter `url` of type `data.UrlQueue`, which represents the URL to be deleted from the queue.
// This function does not return any value.
func (db *Database) DeleteFromQueue(url data.UrlQueue) {
	log.Println("[+] Deleting from queue...", url.URL)

	_, err := db.Collection("url_queues").DeleteOne(context.TODO(), url)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[+] Deleted from queue")
}

// SaveHTML saves the HTML of a crawled page to the database.
//
// page: the crawled page to be saved.
func (db *Database) SaveHTML(page data.CrawledPage) {
	log.Println("[+] Saving html...", page.URL)

	_, err := db.Collection("crawled_pages").InsertOne(context.TODO(), page)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[+] Saved HTML!")
}

// CanQueueUrl checks if a URL can be queued.
//
// Parameters:
// - url: the URL to check.
//
// Returns:
// - bool: true if the URL can be queued, false otherwise.
func (db *Database) CanQueueUrl(url string) bool {
	parsedURL, err := netURL.Parse(url)
	if err != nil {
		log.Fatalln(err)
	}

	existsInQueue := db.Collection("url_queues").FindOne(context.TODO(), bson.D{{Key: "_id", Value: parsedURL.Path}}).Err() == nil
	existsInCrawledPages := db.Collection("crawled_pages").FindOne(context.TODO(), bson.D{{Key: "_id", Value: parsedURL.Path}}) == nil
	existsInIndexedPages := db.Collection("indexed_pages").FindOne(context.TODO(), bson.D{{Key: "_id", Value: parsedURL.Path}}) == nil
	existsInIndexedProducts := db.Collection("indexed_products").FindOne(context.TODO(), bson.D{{Key: "_id", Value: parsedURL.Path}}) == nil

	canQueue := !existsInQueue && !existsInCrawledPages && !existsInIndexedPages && !existsInIndexedProducts

	return canQueue
}

// GetCrawledPages retrieves crawled pages for a given source.
//
// Parameters:
// - source: a string representing the source of the crawled pages. e.g. Jumia
//
// Returns:
// - an array of data.CrawledPage representing the retrieved crawled pages.
func (db *Database) GetCrawledPages(source string) []data.CrawledPage {
	log.Printf("[+] Getting crawled pages for %s...", source)

	res, err := db.Collection("crawled_pages").Find(context.TODO(), bson.D{{Key: "source", Value: source}}, &options.FindOptions{Limit: options.Count().SetLimit(5).Limit})
	if err != nil {
		log.Fatalln(err)
	}

	var pages []data.CrawledPage
	res.All(context.TODO(), &pages)

	log.Printf("[+] Crawled pages for %s retrieved!", source)

	return pages
}

// IndexProduct saves a product to the indexed_products collection in the database.
//
// It takes a parameter `product` of type `data.Product`.
func (db *Database) IndexProduct(product data.Product) {
	log.Println("[+] Saving product...", product.Name)

	parsedURL, err := netURL.Parse(product.URL)
	if err != nil {
		log.Println(err)
	}

	product.Slug = strings.Split(parsedURL.Path, "/")[1]

	_, err = db.Collection("indexed_products").InsertOne(context.TODO(), product, &options.InsertOneOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[+] Product Saved!")
}

// DeleteFromCrawledPages deletes a crawled page from the database.
// And moves it to the indexed pages collection
//
// It takes a parameter of type `data.CrawledPage` which represents the page to be deleted.
func (db *Database) MovePageToIndexed(page data.CrawledPage) {
	log.Println("[+] Moving from crawled pages...", page.URL)

	_, err := db.Collection("indexed_pages").InsertOne(context.TODO(), page, &options.InsertOneOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	db.DeleteCrawledPage(page.URL)
	log.Println("[+] Moved Crawled page!")
}

// DeleteCrawledPage deletes a crawled page from the database.
func (db *Database) DeleteCrawledPage(url string) {
	log.Println("[+] Deleting from crawled pages...", url)

	_, err := db.Collection("crawled_pages").DeleteOne(context.TODO(), bson.D{{Key: "url", Value: url}})
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[+] Deleted Crawled page!")
}
