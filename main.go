package main

import (
	"log"
	"sync"

	"github.com/Cedi-Search/Cedi-Search-Engine/crawler"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/jumia"
	"github.com/joho/godotenv"
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

	jumiaIndexer := jumia.NewIndexer(database)

	wg.Add(1)
	go jumiaIndexer.Index(&wg)

	crawler := crawler.NewCrawler(database)
	crawler.Crawl()

	wg.Wait()
}
