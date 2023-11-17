package ishtari

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
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

// Index indexes the crawled pages for Ishtari.
//
// Parameters:
// - wg: A pointer to a sync.WaitGroup that is used to coordinate the goroutines.
func (il *IndexerImpl) Index(wg *sync.WaitGroup) {
	log.Println("[+] Indexing Ishtari...")

	pages := il.db.GetCrawledPages("Ishtari")

	if len(pages) == 0 {
		log.Println("[+] No pages to index for Ishtari!")
		log.Println("[+] Waiting 60s to continue indexing...")

		time.Sleep(60 * time.Second)

		il.Index(wg)

		wg.Done()
		return
	}

	for _, page := range pages {
		parsedPage := soup.HTMLParse(page.HTML)

		log.Println(page.URL)

		productName := parsedPage.Find("h1", "class", "text-d22").Text()

		productPriceStirng := strings.ReplaceAll(parsedPage.Find("span", "class", "false").Text(), " GH¢", "")

		price, err := strconv.ParseFloat(strings.ReplaceAll(productPriceStirng, ",", ""), 64)

		if err != nil {
			log.Fatalln(err)
		}

		productDescription := parsedPage.Find("div", "class", "my-content").FullText()

		productID := uuid.New()

		productImagesEl := parsedPage.FindAll("img", "class", "border-dgreyZoom")

		productImages := []string{}

		for _, el := range productImagesEl {
			productImages = append(productImages, el.Attrs()["src"])
		}

		productData := data.Product{
			Name:        productName,
			Price:       price,
			Rating:      0,
			Description: productDescription,
			URL:         page.URL,
			Source:      page.Source,
			ProductID:   productID.String(),
			Images:      productImages,
		}

		il.db.IndexProduct(productData)
		il.db.MovePageToIndexed(page)

	}

	il.Index(wg)

}
