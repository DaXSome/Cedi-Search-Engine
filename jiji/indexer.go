package jiji

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	database "github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/models"
	"github.com/anaskhan96/soup"
)

type Indexer interface{}

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

// Index indexes the crawled pages for Jiji.
//
// Parameters:
// - wg: A pointer to a sync.WaitGroup that is used to coordinate the goroutines.
func (il *IndexerImpl) Index(wg *sync.WaitGroup) {
	log.Println("[+] Indexing Jiji...")

	pages := il.db.GetCrawledPages("Jiji")

	if len(pages) == 0 {
		log.Println("[+] No pages to index for Jiji!")
		log.Println("[+] Waiting 60s to continue indexing...")

		time.Sleep(60 * time.Second)

		il.Index(wg)

		wg.Done()
		return
	}

	for _, page := range pages {
		parsedPage := soup.HTMLParse(page.HTML)

		// E.g Kia Sorento 2.5 D Automatic 2003 Red in Akuapim South - Cars, Gabriel Sokah | Jiji.com.gh
		productName := parsedPage.Find("title").Text()

		productName = strings.Split(productName, " in ")[0]

		productPriceEl := parsedPage.Find("span", "itemprop", "price")

		if productPriceEl.Error != nil {
			continue
		}

		productPriceString := productPriceEl.Attrs()["content"]

		price, err := strconv.ParseFloat(productPriceString, 64)

		if err != nil {
			log.Fatalln(err)
		}

		productDescription := parsedPage.Find("span", "class", "qa-description-text").Text()

		productIDParts := strings.Split(page.URL, "-")
		productID := strings.ReplaceAll(productIDParts[len(productIDParts)-1], ".html", "")

		productImagesEl := parsedPage.FindAll("img", "class", "qa-carousel-thumbnail__image")

		productImages := []string{}

		for _, el := range productImagesEl {
			productImages = append(productImages, el.Attrs()["src"])
		}

		if len(productImages) == 0 {
			productImages = append(productImages, parsedPage.Find("img", "class", "b-slider-image").Attrs()["src"])
		}

		productData := models.Product{
			Name:        productName,
			Price:       price,
			Rating:      0,
			Description: productDescription,
			URL:         page.URL,
			Source:      page.Source,
			ProductID:   productID,
			Images:      productImages,
		}

		il.db.IndexProduct(productData)
		il.db.DeleteFromCrawledPages(page)

	}

	il.Index(wg)

}
