package sniffer

import (
	"log"
	"net/url"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"
	"github.com/anaskhan96/soup"
)

func Sniff(target data.Target, db *database.Database) {
	utils.Logger(utils.Sniffer, target.Target, "Sniffing...")

	link := url.URL{}

	link.Host = target.Host
	link.Path = target.SeedPath
	link.Scheme = "https"

	resp := utils.FetchPage(link.String(), "rod")

	doc := soup.HTMLParse(resp)

	links := doc.FindAll("a")

	utils.ShuffleLinks(links)

	newlyFound := []data.Target{}

	for _, link := range links {
		categoryLink := link.Attrs()["href"]

		u, err := url.Parse(categoryLink)
		if err != nil {
			log.Println(err)
			return
		}

		if u.Host != target.Host && u.Host != "" {
			continue
		}

		if u.Host == "" {
			u.Host = target.Host
			u.Scheme = "https"
		}

		if canQueue, err := db.CanQueueUrl(u.String()); err == nil && canQueue {

			newTarget := data.Target{
				Target:   target.Target,
				Host:     target.Host,
				SeedPath: u.Path,
			}

			db.AddToQueue(data.UrlQueue{
				URL:    u.String(),
				Source: target.Target,
			})

			targetExists := false

			for _, data := range newlyFound {
				if data.SeedPath == newTarget.SeedPath {
					targetExists = true
					break
				}
			}

			if !targetExists {
				newlyFound = append(newlyFound, newTarget)
			}

		}

	}

	utils.Logger(utils.Sniffer, target.Target, "Wait 120s to continue sniff")
	time.Sleep(120 * time.Second)

	for _, newTarget := range newlyFound {

		go Sniff(newTarget, db)

	}
}
