package crawler

import (
	"fmt"
	"log"
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
func (cr *Crawler) Crawl(source string) {
	queue, err := cr.db.GetQueue(source)
	if utils.HandleErr(err, "Failed to get pages for crawler") {
		return
	}

	if len(queue) == 0 {
		log.Println("[+] Queue is empty!")
		return
	}

	wg := sync.WaitGroup{}

	for _, url := range queue {

		wg.Add(1)
		go func(url data.UrlQueue) {
			defer wg.Done()

			log.Println("[+] Crawling: ", url.URL)

			var resp string

			if url.Source == "Jiji" || url.Source == "Deus" {
				resp = utils.FetchPage(url.URL, "soup")
			} else {
				resp = utils.FetchPage(url.URL, "rod")
			}

			doc := soup.HTMLParse(resp)

			err := cr.db.SaveHTML(data.CrawledPage{
				URL:    url.URL,
				HTML:   doc.HTML(),
				Source: url.Source,
			})

			if utils.HandleErr(err, fmt.Sprintf("Failed to save HTML: %v", url)) {
				return
			}

			err = cr.db.DeleteFromQueue(url)

			if utils.HandleErr(err, fmt.Sprintf("Failed to delete from crawler queue: %v", url)) {
				return
			}

			log.Println("[+] Crawled: ", url.URL)
		}(url)
	}

	wg.Wait()

	log.Println("[+] Wait 30s to continue crawling")
	time.Sleep(30 * time.Second)

	cr.Crawl(source)
}
