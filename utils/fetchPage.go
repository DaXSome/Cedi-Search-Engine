package utils

import (
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

var browser = rod.New().MustConnect().WithPanic(func(i interface{}) {
	log.Println("[!] Headerless browser proberly lost context.")
})

// FetchPage fetches the content of a web page given its URL.
//
// href: The URL of the web page to fetch.
func FetchPage(href string) string {

	page := browser.MustPage()

	defer page.Close()

	page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "cedisearchbot/0.1 (+https://cedi-search.vercel.app/about)",
	})

	page.Navigate(href)

	html := page.MustWaitStable().MustHTML()

	return html
}
