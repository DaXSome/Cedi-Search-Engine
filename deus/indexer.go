package deus

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	database "github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/anaskhan96/soup"
	"github.com/google/uuid"
)

type IndexerImpl struct {
	db *database.Database
}

// NewIndexer creates a new instance of IndexerImpl.
//
// It takes a database as a parameter and returns a pointer to IndexerImpl.
func NewIndexer(database *database.Database) *IndexerImpl {
	return &IndexerImpl{
		db: database,
	}
}

// Index indexes the crawled pages for Deus.
//
// Parameters:
// - wg: A pointer to a sync.WaitGroup that is used to coordinate the goroutines.
func (il *IndexerImpl) Index(wg *sync.WaitGroup) {
	log.Println("[+] Indexing Deus...")

	pages := il.db.GetCrawledPages("Deus")

	if len(pages) == 0 {
		log.Println("[+] No pages to index for Deus!")
		log.Println("[+] Waiting 60s to continue indexing...")

		time.Sleep(60 * time.Second)

		il.Index(wg)

		wg.Done()
		return
	}

	for _, page := range pages {
		parsedPage := soup.HTMLParse(page.HTML)

		productNameEl := parsedPage.Find("span", "itemprop", "name")

		if productNameEl.Error != nil {
			il.db.DeleteCrawledPage(page.URL)
			continue
		}

		productName := productNameEl.Text()

		productPriceStirng := parsedPage.Find("span", "data-price-type", "finalPrice").Attrs()["data-price-amount"]

		price, err := strconv.ParseFloat(productPriceStirng, 64)

		if err != nil {
			log.Fatalln(err)
		}

		productDescription := parsedPage.Find("div", "class", "description").FullText()

		productID := uuid.New()

		productImage := parsedPage.Find("img", "class", "no-sirv-lazy-load").Attrs()["src"]

		productData := data.Product{
			Name:        productName,
			Price:       price,
			Rating:      0,
			Description: productDescription,
			URL:         page.URL,
			Source:      page.Source,
			ProductID:   productID.String(),
			Images:      []string{productImage},
		}

		il.db.IndexProduct(productData)
		il.db.MovePageToIndexed(page)

	}

	il.Index(wg)

}
