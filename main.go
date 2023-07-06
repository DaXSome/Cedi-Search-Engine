package main

import (
	"github.com/joho/godotenv"
)

var (
	SEED_URLS = []URL{
		{
			URL: "https://www.jumia.com.gh",
			ItemTag: ItemTag{
				Attr:        ".core",
				ValuePrefix: "https://www.jumia.com.gh",
			},
		},
		{
			URL: "https://www.jumia.com.gh/smartphones/",
			ItemTag: ItemTag{
				Attr:        ".core",
				ValuePrefix: "https://www.jumia.com.gh",
			},
		},
		{
			URL: "https://www.jumia.com.gh/hot-beverages-coffee-tea-cocoa/",
			ItemTag: ItemTag{
				Attr:        ".core",
				ValuePrefix: "https://www.jumia.com.gh",
			},
		},
	}
)

func main() {
	godotenv.Load()

	database := DatabaseImpl{}

	database.Init()

	queue := database.GetURLQueues()

	if len(queue) == 0 {
		for _, url := range SEED_URLS {
			queue = append(queue, URLQueue{
				URL: url,
			})
		}
	}

	crawler := CrawlerImpl{}
	// renderer := RendererImpl{}

	crawler.Crawl(queue)
	// renderer.HandleQueue(queue)
}
