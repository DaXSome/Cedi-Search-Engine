package crawler

import (
	"log"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/models"
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
func (cr *Crawler) Crawl() {
	queue := cr.db.GetQueue()

	if len(queue) == 0 {
		log.Println("[+] Queue is empty!")
		return
	}

	wg := sync.WaitGroup{}

	for _, url := range queue {

		wg.Add(1)
		go func(url models.UrlQueue) {

			log.Println("[+] Crawling: ", url.URL)

			soup.Header("User-Agent", "cedisearchbot/0.1 (+https://cedi-search.vercel.app/about)")

			resp := utils.FetchPage(url.URL)

			doc := soup.HTMLParse(resp)

			cr.db.SaveHTML(models.CrawledPage{
				URL:    url.URL,
				HTML:   doc.HTML(),
				Source: url.Source,
			})

			cr.db.DeleteFromQueue(url)

			log.Println("[+] Crawled: ", url.URL)

			wg.Done()

		}(url)
	}

	wg.Wait()

	log.Println("[+] Wait 30s to continue crawling")
	time.Sleep(30 * time.Second)

	cr.Crawl()
}
