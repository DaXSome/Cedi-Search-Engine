package oraimo

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"

	"github.com/anaskhan96/soup"
)

// queueProducts processes a list of products and adds eligible URLs to the queue.
//
// It takes a pointer to a database object 'db' and a slice of 'products' which is a collection of soup.Root objects.
// The function iterates over each 'link' in 'products' and generates a product link.
// If the generated product link is eligible to be queued, it adds it to the database queue using 'db.AddToQueue'.
func queueProducts(db *database.Database, products []soup.Root) {
	for _, link := range products {
		// E.g. https://gh.oraimo.com/oraimo-freepods-lite-40-hour-playtime-enc-true-wireless-earbuds.html
		productLink := link.Attrs()["href"]

		if db.CanQueueUrl(productLink) {
			db.AddToQueue(data.UrlQueue{
				URL:    productLink,
				Source: "Oraimo",
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

	resp := utils.FetchPage(href, "rod")

	doc := soup.HTMLParse(resp)

	return doc.FindAll("a", "class", "product")
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

// Sniff sniffs the website "https://gh.oraimo.com/"
// and extracts link pages to be crawled later
//
// The function takes a pointer to a sync.WaitGroup as a parameter.
func (sl *SnifferImpl) Sniff(wg *sync.WaitGroup) {
	log.Println("[+] Sniffing...")

	defer wg.Done()

	resp := utils.FetchPage("https://gh.oraimo.com/", "rod")

	doc := soup.HTMLParse(resp)

	links := doc.FindAll("a", "role", "menuitem")

	utils.ShuffleLinks(links)

	for _, link := range links {
		// E.g. https://gh.oraimo.com/products/lifestyle/electric-toothbrush.html
		categoryLink := link.Attrs()["href"]

		if strings.Contains(categoryLink, "products") {

			products := extractProducts(categoryLink)

			queueProducts(sl.db, products)

			log.Println("[+] Wait 15s to continue sniff")
			time.Sleep(15 * time.Second)
		}

	}
}
