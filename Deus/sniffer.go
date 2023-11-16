package deus

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/models"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"

	"github.com/anaskhan96/soup"
)

// ShuffleLinks shuffles the order of links.
func shuffleLinks(links []soup.Root) {
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
		productLink := link.Attrs()["href"]

		if db.CanQueueUrl(productLink) {
			db.AddToQueue(models.UrlQueue{
				URL:    productLink,
				Source: "Deus",
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
// The function returns a slice of soup.Root. The slice of
// soup.Root contains the extracted products. The integer represents the total
// number of pages of products.
func extractProducts(href string) []soup.Root {
	log.Println("[+] Extracting products from", href)

	resp := utils.FetchPage(href)

	doc := soup.HTMLParse(resp)

	return doc.FindAll("a", "class", "product-item-photo")
}

type Sniffer interface {
	Sniff(wg *sync.WaitGroup)
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

// Sniff sniffs the website "https://www.deus.com.gh"
// and extracts link pages to be crawled later
//
// The function takes a pointer to a sync.WaitGroup as a parameter.
func (sl *SnifferImpl) Sniff(wg *sync.WaitGroup) {
	log.Println("[+] Sniffing...")

	defer wg.Done()

	resp, err := soup.Get("https://deus.com.gh/")

	if err != nil {
		log.Fatalln(err)
	}

	doc := soup.HTMLParse(resp)

	links := doc.FindAll("a", "class", "child-cat-a")

	shuffleLinks(links)

	for _, link := range links {
		// E.g. https://deus.com.gh/shop/printer-supplies/epson.html
		categoryLink := link.Attrs()["href"]

		products := extractProducts(categoryLink)

		queueProducts(sl.db, products)

		log.Println("[+] Wait 30s to continue sniff")
		time.Sleep(30 * time.Second)

	}
}
