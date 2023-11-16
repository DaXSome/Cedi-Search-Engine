package database

import (
	"context"
	"log"
	"os"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

// openCollection opens a collection in the database.
//
// It takes a pointer to a Database struct and the name of the collection as parameters.
// It returns a driver.Collection.
func openCollection(db *Database, collection string) driver.Collection {
	database, err := db.client.Database(context.Background(), "cedi_search")

	if err != nil {
		log.Fatalln(err)
	}

	col, err := database.Collection(context.TODO(), collection)

	if err != nil {
		log.Fatalln(err)
	}

	return col

}

type Database struct {
	client driver.Client
}

// NewDatabase initializes a new instance of the Database struct.
//
// Returns a pointer to the newly created Database.
func NewDatabase() *Database {
	return &Database{}
}

// Init initializes the database.
//
// It establishes a connection to the database using the provided connection string
// and authentication credentials. If successful, it sets the client field of the
// Database struct to the newly created client.
func (db *Database) Init() {
	log.Println("[+] Initing database...")

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{os.Getenv("DB_CONNECTION_STRING")},
	})

	if err != nil {
		log.Fatalln(err)
	}

	db.client, err = driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD")),
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[+] Database initialized!")

}

// GetQueue retrieves a slice of data.UrlQueue from the Database.
// It randomly selects 10 URLs from the queue and returns
// them as a slice of data.UrlQueue.
func (db *Database) GetQueue() []data.UrlQueue {

	log.Println("[+] Getting queue...")

	ctx := context.Background()
	query := `FOR d IN url_queues
				LET randomValue = RAND()
        		SORT randomValue ASC
				LIMIT 10 
				RETURN d
			`
	database, err := db.client.Database(ctx, "cedi_search")

	if err != nil {
		log.Fatalln(err)
	}

	cursor, err := database.Query(ctx, query, nil)

	if err != nil {
		log.Fatalln(err)
	}

	defer cursor.Close()

	queue := []data.UrlQueue{}

	for {

		var doc data.UrlQueue

		_, err := cursor.ReadDocument(ctx, &doc)

		if doc.URL != "" {
			queue = append(queue, doc)
		}

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

	}

	log.Println("[+] Queue retrieved!")

	return queue
}

// AddToQueue adds a URL to the queue in the Database.
//
// It takes a parameter 'url' of type `data.UrlQueue` which represents the URL to be added.
func (db *Database) AddToQueue(url data.UrlQueue) {
	log.Println("[+] Adding to queue...", url.URL)

	col := openCollection(db, "url_queues")

	_, err := col.CreateDocument(context.TODO(), url)

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

	ctx := context.Background()
	query := `FOR d IN url_queues 
			FILTER d.url == @url
			REMOVE d IN url_queues
			RETURN OLD
			`

	database, err := db.client.Database(ctx, "cedi_search")

	if err != nil {
		log.Fatalln(err)
	}

	bindVars := map[string]interface{}{
		"url": url.URL,
	}

	cursor, err := database.Query(ctx, query, bindVars)

	if err != nil {
		log.Fatalln(err)
	}

	cursor.Close()

	log.Println("[+] Deleted from queue")
}

// SaveHTML saves the HTML of a crawled page to the database.
//
// page: the crawled page to be saved.
func (db *Database) SaveHTML(page data.CrawledPage) {
	log.Println("[+] Saving html...", page.URL)

	col := openCollection(db, "crawled_pages")

	_, err := col.CreateDocument(context.TODO(), page)

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
	ctx := driver.WithQueryCount(context.Background())
	query := `FOR d IN url_queues 
				FILTER d.url == @url
				RETURN d`

	database, err := db.client.Database(ctx, "cedi_search")

	if err != nil {
		log.Fatalln(err)
	}

	bindVars := map[string]interface{}{
		"url": url,
	}

	urlQueuesCursor, err := database.Query(ctx, query, bindVars)

	if err != nil {
		log.Fatalln(err)
	}

	defer urlQueuesCursor.Close()

	query = `FOR d IN crawled_pages 
				FILTER d.url == @url
				RETURN d`

	crawledPagesCursor, err := database.Query(ctx, query, bindVars)

	if err != nil {
		log.Fatalln(err)
	}

	defer crawledPagesCursor.Close()

	query = `FOR d IN indexed_pages 
				FILTER d.url == @url
				RETURN d`

	indexedPagesCursor, err := database.Query(ctx, query, bindVars)

	if err != nil {
		log.Fatalln(err)
	}

	defer indexedPagesCursor.Close()

	return urlQueuesCursor.Count() == 0 && crawledPagesCursor.Count() == 0 && indexedPagesCursor.Count() == 0

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

	ctx := context.Background()
	query := `FOR d IN crawled_pages
				FILTER d.source == @source
				LET randomValue = RAND()
				SORT randomValue ASC
				LIMIT 5
				RETURN d
			`
	database, err := db.client.Database(ctx, "cedi_search")

	if err != nil {
		log.Fatalln(err)
	}

	bindVars := map[string]interface{}{
		"source": source,
	}

	cursor, err := database.Query(ctx, query, bindVars)

	if err != nil {
		log.Fatalln(err)
	}

	defer cursor.Close()

	pages := []data.CrawledPage{}

	for {

		var doc data.CrawledPage

		_, err := cursor.ReadDocument(ctx, &doc)

		if doc.URL != "" {
			pages = append(pages, doc)
		}

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

	}

	log.Printf("[+] Crawled pages for %s retrieved!", source)

	return pages
}

// IndexProduct saves a product to the indexed_products collection in the database.
//
// It takes a parameter `product` of type `data.Product`.
func (db *Database) IndexProduct(product data.Product) {
	log.Println("[+] Saving product...", product.Name)

	col := openCollection(db, "indexed_products")

	_, err := col.CreateDocument(context.TODO(), product)

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

	col := openCollection(db, "indexed_pages")

	_, err := col.CreateDocument(context.TODO(), page)

	if err != nil {
		log.Fatalln(err)
	}

	db.DeleteCrawledPage(page.URL)

	log.Println("[+] Moved Crawled page!")

}

// DeleteCrawledPage deletes a crawled page from the database.
func (db *Database) DeleteCrawledPage(url string) {
	log.Println("[+] Deleting from crawled pages...", url)

	database, err := db.client.Database(context.Background(), "cedi_search")

	if err != nil {
		log.Fatalln(err)
	}

	query := `FOR d IN crawled_pages 
			FILTER d.url == @url
			REMOVE d IN crawled_pages
			RETURN OLD
			`

	if err != nil {
		log.Fatalln(err)
	}

	bindVars := map[string]interface{}{
		"url": url,
	}

	cursor, err := database.Query(context.Background(), query, bindVars)

	if err != nil {
		log.Fatalln(err)
	}

	cursor.Close()

	log.Println("[+] Deleted Crawled page!")

}
