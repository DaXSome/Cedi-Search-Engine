package jumia

import (
	"fmt"
	"log"
	"strconv"
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
		// E.g. https://www.jumia.com.gh/jameson-irish-whiskey-750ml-51665215.html
		productLink := fmt.Sprintf("https://www.jumia.com.gh%s", link.Attrs()["href"])

		if db.CanQueueUrl(productLink) {
			db.AddToQueue(data.UrlQueue{
				URL:    productLink,
				Source: "Jumia",
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

	resp := utils.FetchPage(href)

	doc := soup.HTMLParse(resp)

	totalPagesEl := doc.FindAll("a", "class", "pg")

	totalPages := 0

	if len(totalPagesEl) > 0 {
		lastPageLink := totalPagesEl[len(totalPagesEl)-1].Attrs()["href"]

		eqSignSplit := strings.Split(lastPageLink, "=")

		var err error
		if len(eqSignSplit) > 1 {
			totalPages, err = strconv.Atoi(strings.Split(eqSignSplit[1], "#")[0])
			if err != nil {
				log.Println(err)
			}
		}
	}

	return doc.FindAll("a", "class", "core"), totalPages
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

	resp := utils.FetchPage("https://www.jumia.com.gh")

	doc := soup.HTMLParse(resp)

	links := doc.FindAll("a", "role", "menuitem")

	utils.ShuffleLinks(links)

	for _, link := range links {
		// E.g. https://www.jumia.com.gh/groceries
		categoryLink := link.Attrs()["href"]

		if categoryLink != "" {
			if !strings.Contains(categoryLink, "jumia") {
				categoryLink = fmt.Sprintf("https://www.jumia.com.gh%s", categoryLink)
			}

			products, totalPages := extractProducts(categoryLink)

			queueProducts(sl.db, products)

			for i := 2; i <= totalPages; i++ {
				go func(i int) {
					// E.g. https://www.jumia.com.gh/groceries?page=2
					pageLink := fmt.Sprintf("%s?page=%d", categoryLink, i)

					pageProducts, _ := extractProducts(pageLink)

					queueProducts(sl.db, pageProducts)
				}(i)

			}

			log.Println("[+] Wait 120s to continue sniff")
			time.Sleep(120 * time.Second)

		}
	}
}
