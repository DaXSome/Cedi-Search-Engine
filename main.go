package main

import (
	"log"
	"sync"

	"github.com/Cedi-Search/Cedi-Search-Engine/crawler"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/jiji"
	"github.com/Cedi-Search/Cedi-Search-Engine/jumia"
	"github.com/anaskhan96/soup"
	"github.com/joho/godotenv"
)

func main() {

	log.Println("[+] Startup")

	soup.Header("User-Agent", "cedisearchbot/0.1 (+https://cedi-search.vercel.app/about)")

	wg := sync.WaitGroup{}

	godotenv.Load()

	database := database.NewDatabase()

	database.Init()

	database.UploadProducts()

	jumiaSniffer := jumia.NewSniffer(database)
	jijiSniffer := jiji.NewSniffer(database)

	jumiaIndexer := jumia.NewIndexer(database)

	wg.Add(1)
	go jumiaSniffer.Sniff(&wg)

	wg.Add(1)
	go jijiSniffer.Sniff(&wg)

	wg.Add(1)
	go jumiaIndexer.Index(&wg)

	crawler := crawler.NewCrawler(database)
	crawler.Crawl()

	wg.Wait()
}
