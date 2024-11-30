package main

import (
	"sync"

	"github.com/Cedi-Search/Cedi-Search-Engine/config"
	"github.com/Cedi-Search/Cedi-Search-Engine/crawler"
	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/deus"
	"github.com/Cedi-Search/Cedi-Search-Engine/ishtari"
	"github.com/Cedi-Search/Cedi-Search-Engine/jiji"
	"github.com/Cedi-Search/Cedi-Search-Engine/jumia"
	"github.com/Cedi-Search/Cedi-Search-Engine/oraimo"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"
	"github.com/anaskhan96/soup"
	"github.com/joho/godotenv"
)

func main() {
	utils.Logger(utils.Default, utils.Default, "Startup")

	soup.Header("User-Agent",config.USER_AGENT)

	wg := sync.WaitGroup{}

	godotenv.Load()

	database := database.NewDatabase()

	crawler := crawler.NewCrawler(database)

	targets := []data.Target{
		deus.NewDeus(database),
		jumia.NewJumia(database),
		jiji.NewJiji(database),
		ishtari.NewIshtari(database),
		oraimo.NewOraimo(database),
	}

	wg.Add(len(targets))
	for _, target := range targets {
		go target.Sniff(&wg)
		go crawler.Crawl(target.String(), target.Index)
	}

	wg.Wait()
}
