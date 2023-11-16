package utils

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// FetchPage fetches the content of a web page given its URL.
//
// href: The URL of the web page to fetch.
func FetchPage(href string) string {

	browser := rod.New().MustConnect()

	page := browser.MustPage()

	page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "cedisearchbot/0.1 (+https://cedi-search.vercel.app/about)",
	})

	page.Navigate(href)

	html := page.MustWaitStable().MustHTML()

	return html
}
