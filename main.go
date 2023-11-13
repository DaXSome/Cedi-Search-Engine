package main

import (
	"log"
	"sync"

	"github.com/joho/godotenv"
	"github.com/owbird/cedisearch/crawler"
	"github.com/owbird/cedisearch/database"
	"github.com/owbird/cedisearch/jumia"
)

func main() {

	log.Println("[+] Startup")

	wg := sync.WaitGroup{}

	godotenv.Load()

	database := database.NewDatabase()

	database.Init()

	jumiaSniffer := jumia.NewSniffer(database)

	wg.Add(1)
	go jumiaSniffer.Sniff(&wg)

	crawler := crawler.NewCrawler(database)
	crawler.Crawl()

	wg.Wait()
}
