package ishtari

import (
	"fmt"
	"log"
	"strconv"
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

		// E.g. https://ishtari.com.gh/USB-Desktop-Microphone-With-Tripod-/p=815
		productLink := fmt.Sprintf("https://ishtari.com.gh%s", link.Attrs()["href"])

		if db.CanQueueUrl(productLink) {
			db.AddToQueue(data.UrlQueue{
				URL:    productLink,
				Source: "Ishtari",
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
func extractProducts(href string) ([]soup.Root, int) {
	log.Println("[+] Extracting products from", href)

	resp := utils.FetchPage(href, "rod")

	doc := soup.HTMLParse(resp)

	paginationEl := doc.Find("ul", "class", "category-pagination")

	totalPages := 0

	var err error

	if paginationEl.Error == nil {
		paginationChildren := paginationEl.Children()

		totalPages, err = strconv.Atoi(paginationChildren[len(paginationChildren)-2].FullText())

		if err != nil {
			log.Fatalln(err)
		}

	}

	return doc.FindAll("a", "class", "false"), totalPages
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

// Sniff sniffs the website "https://ishtari.com.gh/"
// and extracts link pages to be crawled later
//
// The function takes a pointer to a sync.WaitGroup as a parameter.
func (sl *SnifferImpl) Sniff(wg *sync.WaitGroup) {
	log.Println("[+] Sniffing...")

	defer wg.Done()

	html := utils.FetchPage("https://ishtari.com.gh/", "rod")

	doc := soup.HTMLParse(html)

	links := doc.FindAll("a", "class", "text-d13")

	utils.ShuffleLinks(links)

	for _, link := range links {
		// E.g. /Electronics/c=1023
		categoryLink := link.Attrs()["href"]

		categoryLink = fmt.Sprintf("https://ishtari.com.gh%s", categoryLink)

		products, totalPages := extractProducts(categoryLink)

		queueProducts(sl.db, products)

		for i := 2; i <= totalPages; i++ {
			go func(i int) {
				// E.g. https://ishtari.com.gh/Back-To-School/c=918?page=6
				pageLink := fmt.Sprintf("%s?page=%d", categoryLink, i)

				pageProducts, _ := extractProducts(pageLink)

				queueProducts(sl.db, pageProducts)
			}(i)

		}

		log.Println("[+] Wait 120s to continue sniff")
		time.Sleep(120 * time.Second)

	}
}
