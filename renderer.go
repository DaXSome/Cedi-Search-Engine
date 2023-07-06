package main

import (
	"context"
	"log"
	"sync"

	"github.com/chromedp/chromedp"
)

type Renderer interface {
	ReadURL(url string)
}

type RendererImpl struct {
	database DatabaseImpl
}

func (rl *RendererImpl) Save(html string, url string) error {
	log.Println("SAVING", url)

	rl.database.SaveRawData(html, url)
	return nil
}

func (rl *RendererImpl) ReadURL(url string, wg *sync.WaitGroup, counter chan struct{}) (string, error) {
	counter <- struct{}{}

	log.Println("READING", url)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("cedisearchbot/0.1 (+http://www.cedisearch.com/bot.html)"),
	)

	exec_ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	var html string

	ctx, cancel := chromedp.NewContext(exec_ctx)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	)

	if err != nil {
		log.Fatal(err)
	}

	rl.Save(html, url)

	wg.Done()
	<-counter
	return "", nil
}

func (rl *RendererImpl) HandleQueue(urls []string) {
	counter := make(chan struct{}, 5)

	wg := sync.WaitGroup{}

	for _, url := range urls {

		wg.Add(1)
		go rl.ReadURL(url, &wg, counter)

	}

	wg.Wait()
}
