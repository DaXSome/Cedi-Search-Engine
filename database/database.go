package database

import (
	"context"
	"encoding/json"
	"log"
	"os"

	models "github.com/Cedi-Search/Cedi-Search-Engine/models"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
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

// GetQueue retrieves a slice of models.UrlQueue from the Database.
// It randomly selects 10 URLs from the queue and returns
// them as a slice of models.UrlQueue.
func (db *Database) GetQueue() []models.UrlQueue {

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

	queue := []models.UrlQueue{}

	for {

		var doc models.UrlQueue

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
// It takes a parameter 'url' of type `models.UrlQueue` which represents the URL to be added.
func (db *Database) AddToQueue(url models.UrlQueue) {
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
// It takes a parameter `url` of type `models.UrlQueue`, which represents the URL to be deleted from the queue.
// This function does not return any value.
func (db *Database) DeleteFromQueue(url models.UrlQueue) {
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
func (db *Database) SaveHTML(page models.CrawledPage) {
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
// - an array of models.CrawledPage representing the retrieved crawled pages.
func (db *Database) GetCrawledPages(source string) []models.CrawledPage {

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

	pages := []models.CrawledPage{}

	for {

		var doc models.CrawledPage

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
// It takes a parameter `product` of type `models.Product`.
func (db *Database) IndexProduct(product models.Product) {
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
// It takes a parameter of type `models.CrawledPage` which represents the page to be deleted.
func (db *Database) MovePageToIndexed(page models.CrawledPage) {
	log.Println("[+] Moving from crawled pages...", page.URL)

	col := openCollection(db, "indexed_pages")

	_, err := col.CreateDocument(context.TODO(), page)

	if err != nil {
		log.Fatalln(err)
	}

	db.DeleteCrawledPage(page.URL)

	log.Println("[+] Moved Crawled page!")

}

func (db *Database) UploadProducts() {
	log.Println("[+] Preparing to upload db")

	ctx := context.Background()
	query := `FOR d IN indexed_products
			RETURN {
				product_id: d.product_id,
				name: d.name,
				price: d.price,
				rating: d.rating,
				description: d.description,
				url: d.url,
				source: d.source,
				images: d.images
			}
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

	products := []models.AlgoliaData{}

	searchSuggestionsIndex := []string{}

	for {

		var doc models.Product

		_, err := cursor.ReadDocument(ctx, &doc)

		if doc.URL != "" {
			products = append(products, models.AlgoliaData{
				ObjectID: doc.ProductID,
				Product:  doc,
			})

			searchSuggestionsIndex = append(searchSuggestionsIndex, doc.Name)
		}

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

	}

	client := search.NewClient(os.Getenv("ALGOLIA_APP_ID"), os.Getenv("ALGOLIA_API_KEY"))

	index := client.InitIndex("products")

	_, err = index.SaveObjects(products)

	if err != nil {
		log.Fatalln(err)
	}

	suggestionsIndexJson, err := json.MarshalIndent(searchSuggestionsIndex, "", "")

	if err != nil {
		log.Fatalln(err)
	}

	err = os.WriteFile("index.json", suggestionsIndexJson, 0644)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[+] Uploaded products")

}

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
