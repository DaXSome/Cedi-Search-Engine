package main

import (
	"log"
	"sync"

	"github.com/joho/godotenv"
)

func main() {

	log.Println("[+] Startup")

	wg := sync.WaitGroup{}

	godotenv.Load()

	database := NewDatabase()

	database.Init()

	sniffer := NewSniffer(database)

	wg.Add(1)
	go sniffer.Sniff(&wg)

	crawler := NewCrawler(database)
	crawler.Crawl()

	wg.Wait()
}
