package utils

import (
	"log"

	"github.com/Cedi-Search/Cedi-Search-Engine/config"
	"github.com/anaskhan96/soup"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

var controlUrl = launcher.New().Headless(true).MustLaunch()

var browser = rod.New().ControlURL(controlUrl).MustConnect().WithPanic(func(i interface{}) {
	log.Println("[!] Headerless browser probably lost context.")
})

// FetchPage fetches the content of a web page given its URL.
//
// href: The URL of the web page to fetch.
func FetchPage(href, fetcher string) string {
	Logger(Utils, Utils, "Fetching ", href, " using ", fetcher)

	var html string

	if fetcher == "rod" {

		page := browser.MustPage()

		defer page.Close()

		page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: config.USER_AGENT,
		})

		page.Navigate(href)

		page.MustWaitLoad()

		html = page.MustHTML()
	} else {
		var err error

		html, err = soup.Get(href)
		if err != nil {
			log.Fatalln(err)
		}

	}

	return html
}
