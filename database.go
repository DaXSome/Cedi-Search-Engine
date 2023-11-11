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
	query := "FOR d IN url_queues LIMIT 5 RETURN d"
	database, err := db.client.Database(ctx, "cedi_search")

	if err != nil {
		log.Fatalln(err)
	}

	cursor, err := database.Query(ctx, query, map[string]interface{}{})

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
	log.Println("[+] Adding to queue...")

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
		log.Println(err)
	}

	log.Println("[+] Added to queue!")

}
