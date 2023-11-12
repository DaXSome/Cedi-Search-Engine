package main

import (
	"context"
	"log"
	"os"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

type Database struct {
	client driver.Client
}

func NewDatabase() *Database {
	return &Database{}
}

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

func (db *Database) GetQueue() []UrlQueue {

	log.Println("[+] Getting queue...")

	ctx := context.Background()
	query := `FOR d IN url_queues 
				LIMIT 5 
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

	queue := []UrlQueue{}

	for {

		var doc UrlQueue

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

func (db *Database) AddToQueue(url UrlQueue) {
	log.Println("[+] Adding to queue...", url.URL)

	database, err := db.client.Database(context.Background(), "cedi_search")

	if err != nil {
		log.Fatalln(err)
	}

	col, err := database.Collection(context.TODO(), "url_queues")

	if err != nil {
		log.Fatalln(err)
	}

	_, err = col.CreateDocument(context.TODO(), url)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[+] Added to queue!")

}

func (db *Database) DeleteFromQueue(url UrlQueue) {
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

func (db *Database) SaveHTML(page CrawledPage) {
	log.Println("[+] Saving html...", page.URL)

	database, err := db.client.Database(context.Background(), "cedi_search")

	if err != nil {
		log.Fatalln(err)
	}

	col, err := database.Collection(context.TODO(), "crawled_pages")

	if err != nil {
		log.Fatalln(err)
	}

	_, err = col.CreateDocument(context.TODO(), page)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("[+] Saved HTML!")

}

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

	return urlQueuesCursor.Count() == 0 && crawledPagesCursor.Count() == 0

}
