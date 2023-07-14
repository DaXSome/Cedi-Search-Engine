package main

import (
	"log"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

type Crawler interface {
	ReadURL(url string)
}

type CrawlerImpl struct {
	database Database
}

func (ci *CrawlerImpl) Save(html string, url string) error {
	log.Println("SAVING", url)

	// rl.database.Save(html, url)
	return nil
}

// Crawl is a function that takes a queue of URLs and crawls each URL,
// searching for new links and adding them to a new queue. It uses a counter
// channel to limit the number of concurrent crawls and a wait group to wait
// for all crawls to finish. It creates a new browser instance using Rod, and
// then navigates to each URL in the queue, waits for the page to load, and
// extracts the href attribute of each element specified by the ItemTag. If a
// valid href is found, it constructs a new URLQueue object and adds it to the
// new_queue. Finally, it recursively calls itself with the new_queue to crawl
// the new links. The function has no parameters and does not return anything.
func (ci *CrawlerImpl) Crawl(queue []URLQueue) {
	counter := make(chan struct{}, 1)
	wg := sync.WaitGroup{}

	for _, url := range queue {
		wg.Add(1)
		go func(url URL) {

			// New url queue for new links found
			// in the current url
			new_queue := []URLQueue{}

			counter <- struct{}{}

			log.Println("[+] READING", url.URLString)

			browser := rod.New().MustConnect()

			page := browser.MustPage(url.URLString)

			page.MustSetExtraHeaders("X-Bot-Agent", "cedisearchbot/0.1 (+http://www.cedisearch.com/bot.html)")

			page.Navigate(url.URLString)
			page.MustWaitLoad()
			page.MustWaitStable()
			page.MustWaitRequestIdle()
			page.MustWaitIdle()
			page.MustWaitElementsMoreThan(url.ItemTag.Attr, 5)

			// Scroll down to the bottom of the page
			// End key doesn't work
			// Simulate multiple space key presses
			// to ensure end is reached so some pages load
			// like JUMIA
			for i := 0; i <= 500; i++ {

				page.Keyboard.Press(input.Space)

				// pause for 5 seconds after every 50 presses
				// to allow loading
				if i%50 == 0 {
					time.Sleep(5 * time.Second)
				}
			}

			nodes := page.MustElements(url.ItemTag.Attr)

			for _, node := range nodes {
				data := node.MustAttribute("href")

				if data != nil {
					log.Println("DATA ==> ", *data)

					item_url := url.ItemTag.ValuePrefix + *data

					new_url := URLQueue{
						ID: item_url,
						URLItem: URL{
							URLString: item_url,
							ItemTag:   url.ItemTag,
						},
					}

					new_queue = append(new_queue, new_url)

					ci.database.QueueURL(new_url)

				}

			}

			log.Printf("[+] Found %v new links\n", len(new_queue))
			ci.database.SaveRawData(page.MustHTML(), url.URLString)
			ci.Crawl(new_queue)

			wg.Done()
			<-counter
		}(url.URLItem)
	}

	wg.Wait()
	// return "", nil
}
