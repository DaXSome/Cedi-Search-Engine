package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/Cedi-Search/Cedi-Search-Engine/config"
	"github.com/Cedi-Search/Cedi-Search-Engine/data"
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

	engineConfig, err := os.OpenFile("engine.json", os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	config := data.Config{}

	json.NewDecoder(engineConfig).Decode(&config)

	db := database.NewDatabase()

	wg.Add(len(config.Targets))
	for _, target := range config.Targets {
		go sniffer.Sniff(target, db)
	}

	wg.Wait()
}
