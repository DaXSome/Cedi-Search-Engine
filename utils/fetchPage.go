package utils

import (
	"log"
	"time"

	"github.com/anaskhan96/soup"
)

// FetchPage fetches the content of a web page given its URL.
//
// href: The URL of the web page to fetch.
func FetchPage(href string) string {

	resp, err := soup.Get(href)

	if err != nil {
		for {
			log.Println("[+] Retrying: ", href)

			resp, err = soup.Get(href)

			if err == nil {
				break
			}

			time.Sleep(10 * time.Second)
		}
	}

	return resp
}
