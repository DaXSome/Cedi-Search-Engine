package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

var (
	SEED_URLS = []URLQueue{
		{
			ID: "https://www.jumia.com.gh/white-label-lemongrass-and-ginger-tea-25-tea-bags-77639230.html",
			URLItem: URL{
				URLString: "https://www.jumia.com.gh/white-label-lemongrass-and-ginger-tea-25-tea-bags-77639230.html",
				ItemTag: ItemTag{
					Attr:        ".core",
					ValuePrefix: "https://www.jumia.com.gh",
				},
			},
		},
		{
			ID: "www.jumia.com.gh",
			URLItem: URL{
				URLString: "https://www.jumia.com.gh",
				ItemTag: ItemTag{
					Attr:        ".core",
					ValuePrefix: "https://www.jumia.com.gh",
				},
			},
		},
		{
			ID: "www.jumia.com.gh/smartphones/",
			URLItem: URL{
				URLString: "https://www.jumia.com.gh/smartphones/",
				ItemTag: ItemTag{
					Attr:        ".core",
					ValuePrefix: "https://www.jumia.com.gh",
				},
			},
		},
		{
			ID: "www.jumia.com.gh/hot-beverages-coffee-tea-cocoa/",

			URLItem: URL{
				URLString: "https://www.jumia.com.gh/hot-beverages-coffee-tea-cocoa/",
				ItemTag: ItemTag{
					Attr:        ".core",
					ValuePrefix: "https://www.jumia.com.gh",
				},
			},
		},
	}
)

func main() {
	godotenv.Load()

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	database := &DatabaseImpl{}

	database.Init()

	queue := database.GetURLQueues()

	if len(queue) == 0 {
		queue = SEED_URLS
	}

	fmt.Printf("[+] %v to crawl\n", len(queue))

	crawler := CrawlerImpl{
		database: database,
	}

	crawler.Crawl(queue)

	// renderer := RendererImpl{}

	// renderer.HandleQueue(queue)
}
