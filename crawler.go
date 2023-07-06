package main

import (
	"log"
	"sync"

	"github.com/go-rod/rod"
)

type Crawler interface {
	ReadURL(url string)
}

type CrawlerImpl struct {
	database DatabaseImpl
}

func (rl *CrawlerImpl) Save(html string, url string) error {
	log.Println("SAVING", url)

	// rl.database.Save(html, url)
	return nil
}

func (rl *CrawlerImpl) Crawl(queue []URLQueue) {
	counter := make(chan struct{}, 3)
	wg := sync.WaitGroup{}

	for _, url := range queue {
		wg.Add(1)
		go func(url URL) {
			new_queue := []URLQueue{}

			counter <- struct{}{}

			log.Println("READING", url)

			browser := rod.New().MustConnect()

			page := browser.MustPage(url.URL)

			page.MustSetExtraHeaders("X-Bot-Agent", "cedisearchbot/0.1 (+http://www.cedisearch.com/bot.html)")

			page.Navigate(url.URL)
			page.MustWaitLoad()

			nodes := page.MustElements(url.ItemTag.Attr)

			for _, node := range nodes {
				data := node.MustAttribute("href")

				if data != nil {

					item_url := url.ItemTag.ValuePrefix + *data

					new_queue = append(new_queue, URLQueue{
						URL: URL{
							URL:     item_url,
							ItemTag: url.ItemTag,
						},
					})
				}

			}

			rl.Crawl(new_queue)

			wg.Done()
			<-counter
		}(url.URL)
	}

	wg.Wait()
	// return "", nil
}
