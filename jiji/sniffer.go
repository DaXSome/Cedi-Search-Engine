package jiji

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"

	"github.com/anaskhan96/soup"
)

// ShuffleLinks shuffles the order of links.
func shuffleLinks(links []string) {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	rand.Shuffle(len(links), func(i, j int) {
		links[i], links[j] = links[j], links[i]
	})
}

// queueProducts processes a list of products and adds eligible URLs to the queue.
//
// It takes a pointer to a database object 'db' and a slice of 'products' which is a collection of soup.Root objects.
// The function iterates over each 'link' in 'products' and generates a product link.
// If the generated product link is eligible to be queued, it adds it to the database queue using 'db.AddToQueue'.
func queueProducts(db *database.Database, products []soup.Root) {
	for _, link := range products {
		// E.g. https://jiji.com.gh/us-embassy-area/commercial-properties/apartments-yZ4tX1iUJB0rSdhAdhf1UA7x.html?page=2&pos=1&cur_pos=1&ads_per_page=23&ads_count=63809&lid=Fmd1TGLFlcaLNkMG&indexPosition=0
		productLink := fmt.Sprintf("https://jiji.com.gh%s", link.Attrs()["href"])
		productLink = strings.Split(productLink, "?")[0]

		if db.CanQueueUrl(productLink) {
			db.AddToQueue(data.UrlQueue{
				URL:    productLink,
				Source: "Jiji",
			})
		} else {
			log.Println("[+] Skipping", productLink)
		}

	}

}

// extractProducts extracts products from a given href.
//
// It takes a string parameter, href, which represents the URL from which the
// products will be extracted.
//
// The function returns a slice of soup.Root and an integer. The slice of
// soup.Root contains the extracted products. The integer represents the total
// number of pages of products.
func extractProducts(href string) []soup.Root {
	log.Println("[+] Extracting products from", href)

	resp := utils.FetchPage(href)

	doc := soup.HTMLParse(resp)

	return doc.FindAll("a", "class", "b-list-advert-base")
}

type SnifferImpl struct {
	db *database.Database
}

// NewSniffer creates a new SnifferImpl instance.
//
// It takes a database as a parameter and returns a pointer to a SnifferImpl struct.
func NewSniffer(database *database.Database) *SnifferImpl {
	return &SnifferImpl{
		db: database,
	}
}

// Sniff sniffs the website "https://www.jumia.com.gh"
// and extracts link pages to be crawled later
//
// The function takes a pointer to a sync.WaitGroup as a parameter.
func (sl *SnifferImpl) Sniff(wg *sync.WaitGroup) {
	log.Println("[+] Sniffing...")

	defer wg.Done()

	categories := []string{
		"vehicles",
		"real-estate",
		"mobile-phones-tablets",
		"electronics",
		"home-garden",
		"health-and-beauty",
		"fashion-and-beauty",
		"hobbies-art-sport",
		"seeking-work-cvs",
		"services",
		"jobs",
		"babies-and-kids",
		"animals-and-pets",
		"agriculture-and-foodstuff",
		"office-and-commercial-equipment-tools",
		"repair-and-construction",
	}

	shuffleLinks(categories)

	for _, category := range categories {

		categoryLink := fmt.Sprintf("https://jiji.com.gh/%s", category)

		for i := 1; i <= 1000; i++ {
			pageLink := fmt.Sprintf("%s?page=%d", categoryLink, i)

			// E.g. https://jiji.com.gh/repair-and-construction?page=992
			pageProducts := extractProducts(pageLink)

			queueProducts(sl.db, pageProducts)

			if i%50 == 0 {
				log.Println("[+] Wait 120s to continue sniff")
				time.Sleep(120 * time.Second)
			}

		}

	}

}
