package crawler

import (
	"fmt"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"
	"github.com/anaskhan96/soup"
)

type Crawler struct {
	db *database.Database
}

// NewCrawler creates a new instance of the Crawler struct.
//
// It takes a pointer to a database.Database object as its parameter.
// It returns a pointer to a Crawler object.
func NewCrawler(database *database.Database) *Crawler {
	return &Crawler{
		db: database,
	}
}

// Crawl performs crawling operation.
//
// It retrieves URLs from the database queue and starts crawling each URL concurrently.
// For each URL, it fetches the page content, parses it, saves the HTML to the database,
// and deletes the URL from the queue. Once all URLs have been crawled, it waits for 30 seconds
// before calling itself recursively to continue the crawling process.
func (cr *Crawler) Crawl(source string, indexer func(page data.CrawledPage)) {
	queue, err := cr.db.GetQueue(source)
	if utils.HandleErr(err, "Failed to get pages for crawler") {
		return
	}

	if len(queue) == 0 {
		utils.Logger("crawler", "[+] Queue is empty!")
		return
	}

	wg := sync.WaitGroup{}

	for _, url := range queue {

		wg.Add(1)
		go func(url data.UrlQueue) {
			defer wg.Done()

			utils.Logger("crawler", "[+] Crawling: ", url.URL)

			var resp string

			if url.Source == "Jiji" || url.Source == "Deus" {
				resp = utils.FetchPage(url.URL, "soup")
			} else {
				resp = utils.FetchPage(url.URL, "rod")
			}

			doc := soup.HTMLParse(resp)

			page := data.CrawledPage{
				URL:    url.URL,
				HTML:   doc.HTML(),
				Source: url.Source,
			}

			indexer(page)

			err = cr.db.DeleteFromQueue(url)

			if utils.HandleErr(err, fmt.Sprintf("Failed to delete from crawler queue: %v", url)) {
				return
			}

			utils.Logger("crawler", "[+] Crawled: ", url.URL)
		}(url)
	}

	wg.Wait()

	utils.Logger("crawler", "[+] Wait 30s to continue crawling")
	time.Sleep(30 * time.Second)

	cr.Crawl(source, indexer)
}
