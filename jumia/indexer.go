package jumia

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	database "github.com/Cedi-Search/Cedi-Search-Engine/database"
	models "github.com/Cedi-Search/Cedi-Search-Engine/models"
	"github.com/anaskhan96/soup"
)

type Indexer interface{}

type IndexerImpl struct {
	db *database.Database
}

func NewIndexer(database *database.Database) *IndexerImpl {
	return &IndexerImpl{
		db: database,
	}
}

func (il *IndexerImpl) Index(wg *sync.WaitGroup) {
	log.Println("[+] Indexing Jumia...")

	pages := il.db.GetCrawledPages("Jumia")

	if len(pages) == 0 {
		log.Println("[+] No pages to index for Jumia!")
		log.Println("[+] Waiting 60s to continue indexing...")

		time.Sleep(60 * time.Second)

		il.Index(wg)

		wg.Done()
		return
	}

	for _, page := range pages {
		parsedPage := soup.HTMLParse(page.HTML)

		productName := parsedPage.Find("h1").Text()

		productPriceStirngEl := parsedPage.Find("span", "class", "-prxs")

		productPriceStirng := ""

		if productPriceStirngEl.Error != nil {
			continue
		} else {
			productPriceStirng = productPriceStirngEl.Text()
		}

		priceParts := strings.Split(productPriceStirng, " ")[1]

		price, err := strconv.ParseFloat(strings.ReplaceAll(priceParts, ",", ""), 64)

		if err != nil {
			log.Fatalln(err)
		}

		productRatingText := parsedPage.Find("div", "class", "stars").Text()

		productRatingString := strings.Split(productRatingText, " ")[0]

		rating, err := strconv.ParseFloat(productRatingString, 64)

		if err != nil {
			log.Fatalln(err)
		}

		productDescriptionEl := parsedPage.Find("div", "class", "-mhm")

		productDescription := ""

		if productDescriptionEl.Error == nil {
			productDescription = productDescriptionEl.FullText()
		}

		productID := ""

		productIDTextEl := parsedPage.Find("li", "class", "-pvxs")

		if productIDTextEl.Error == nil {
			productIDText := productIDTextEl.FullText()
			productID = strings.Split(productIDText, " ")[1]
		}

		productImagesEl := parsedPage.FindAll("img", "class", "-fw")

		productImages := []string{}

		for _, el := range productImagesEl {
			productImages = append(productImages, el.Attrs()["data-src"])
		}

		productData := models.Product{
			Name:        productName,
			Price:       price,
			Rating:      rating,
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
