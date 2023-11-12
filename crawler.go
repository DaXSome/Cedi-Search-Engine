package main

import (
	"log"
	"sync"
	"time"

	"github.com/anaskhan96/soup"
)

type Crawler struct {
	db *Database
}

func NewCrawler(database *Database) *Crawler {
	return &Crawler{
		db: database,
	}
}

func (cr *Crawler) Crawl() {
	queue := cr.db.GetQueue()

	if len(queue) == 0 {
		log.Println("[+] Queue is empty!")
		return
	}

	wg := sync.WaitGroup{}

	for _, url := range queue {

		wg.Add(1)
		go func(url UrlQueue) {

			log.Println("[+] Crawling: ", url.URL)

			soup.Header("User-Agent", "cedisearchbot/0.1 (+http://www.cedisearch.com/bot.html)")

			resp, err := soup.Get(url.URL)

			if err != nil {
				log.Fatalln(err)
			}

			doc := soup.HTMLParse(resp)

			cr.db.SaveHTML(CrawledPage{
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

	log.Println("[+] Wait 60s to continue crawling")
	time.Sleep(60 * time.Second)

	cr.Crawl()
}
