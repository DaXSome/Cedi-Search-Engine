package utils

import (
	"log"
	"time"

	"github.com/anaskhan96/soup"
)

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
