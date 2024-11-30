package deus

import (
	"strconv"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"
	"github.com/google/uuid"

	"github.com/anaskhan96/soup"
)

const (
	source = "Deus"
)

type Deus struct {
	db *database.Database
}

// queueProducts processes a list of products and adds eligible URLs to the queue.
//
// It takes a pointer to a database object 'db' and a slice of 'products' which is a collection of soup.Root objects.
// The function iterates over each 'link' in 'products' and generates a product link.
// If the generated product link is eligible to be queued, it adds it to the database queue using 'db.AddToQueue'.
func queueProducts(db *database.Database, products []soup.Root) {
	for _, link := range products {
		productLink := link.Attrs()["href"]

		canQueue, err := db.CanQueueUrl(productLink)
		if utils.HandleErr(err, "Failed to check Deus product queue") {
			continue
		}

		if canQueue {
			err := db.AddToQueue(data.UrlQueue{
				URL:    productLink,
				Source: "Deus",
			})

			utils.HandleErr(err, "Failed to queue Deus product")
		} else {
			utils.Logger(utils.Sniffer, source, "Skipping", productLink)
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
	utils.Logger(utils.Sniffer, "Extracting products from ", href)

	resp := utils.FetchPage(href, "rod")

	doc := soup.HTMLParse(resp)

	return doc.FindAll("a", "class", "product-item-photo")
}

func NewDeus(db *database.Database) *Deus {
	return &Deus{
		db: db,
	}
}

func (deus *Deus) Index(page data.CrawledPage) {
	utils.Logger(utils.Indexer, source, "Indexing Deus...")

	parsedPage := soup.HTMLParse(page.HTML)

	productNameEl := parsedPage.Find("span", "itemprop", "name")

	if productNameEl.Error != nil {
		return
	}

	productName := productNameEl.Text()

	productPriceStirng := parsedPage.Find("span", "data-price-type", "finalPrice").Attrs()["data-price-amount"]

	price, err := strconv.ParseFloat(productPriceStirng, 64)
	if utils.HandleErr(err, "Failed to converted Deus product price") {
		return
	}

	productDescription := ""

	productDescriptionEl := parsedPage.Find("div", "class", "description")

	if productDescriptionEl.Error == nil {
		productDescription = productDescriptionEl.FullText()
	}

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

	err = deus.db.IndexProduct(productData)
	if utils.HandleErr(err, "Couldn't index Deus product") {
		return
	}
}

func (deus *Deus) Sniff(wg *sync.WaitGroup) {
	utils.Logger(utils.Sniffer, source, "Sniffing...")

	defer wg.Done()

	resp := utils.FetchPage("https://deus.com.gh/", "rod")

	doc := soup.HTMLParse(resp)

	links := doc.FindAll("a", "class", "child-cat-a")

	utils.ShuffleLinks(links)

	for _, link := range links {
		// E.g. https://deus.com.gh/shop/printer-supplies/epson.html
		categoryLink := link.Attrs()["href"]

		products := extractProducts(categoryLink)

		queueProducts(deus.db, products)

		utils.Logger(utils.Sniffer, source, "Wait 30s to continue sniff")
		time.Sleep(30 * time.Second)

	}
}

func (d *Deus) String() string { return source }
