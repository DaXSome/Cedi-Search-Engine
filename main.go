package main

import (
	"log"
	"sync"

	"github.com/Cedi-Search/Cedi-Search-Engine/config"
	"github.com/Cedi-Search/Cedi-Search-Engine/crawler"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/sniffer"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"
	"github.com/anaskhan96/soup"
	"github.com/joho/godotenv"
)

func main() {
	utils.Logger(utils.Default, utils.Default, "Startup")

	soup.Header("User-Agent", config.USER_AGENT)

	wg := sync.WaitGroup{}

	godotenv.Load()

	db := database.NewDatabase()

	crawlerFunc := crawler.NewCrawler(db)

	targets, err := db.GetTargets()
	if err != nil {
		log.Fatalln(err)
	}

	wg.Add(len(targets))
	for _, target := range targets {
		go sniffer.Sniff(target, db)
		go crawlerFunc.Crawl(target)
	}

	wg.Wait()
}
