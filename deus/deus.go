package deus

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"
	"github.com/google/uuid"

	"github.com/anaskhan96/soup"
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

	resp := utils.FetchPage(href, "rod")

	doc := soup.HTMLParse(resp)

	return doc.FindAll("a", "class", "product-item-photo")
}

func NewDeus(db *database.Database) *Deus {
	return &Deus{
		db: db,
	}
}

func (deus *Deus) Index(wg *sync.WaitGroup) {
	log.Println("[+] Indexing Deus...")

	pages, err := deus.db.GetCrawledPages("Deus")
	if utils.HandleErr(err, "Failed to index for Deus") {
		return
	}

	if len(pages) == 0 {
		log.Println("[+] No pages to index for Deus!")
		log.Println("[+] Waiting 60s to continue indexing...")

		time.Sleep(60 * time.Second)

		deus.Index(wg)

		wg.Done()
		return
	}

	for _, page := range pages {
		parsedPage := soup.HTMLParse(page.HTML)

		productNameEl := parsedPage.Find("span", "itemprop", "name")

		if productNameEl.Error != nil {
			err := deus.db.DeleteCrawledPage(page.URL)
			utils.HandleErr(err, "Failed to delete Deus crawled page")
			continue
		}

		productName := productNameEl.Text()

		productPriceStirng := parsedPage.Find("span", "data-price-type", "finalPrice").Attrs()["data-price-amount"]

		price, err := strconv.ParseFloat(productPriceStirng, 64)
		if err != nil {
			log.Fatalln(err)
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

		err = deus.db.MovePageToIndexed(page)
		utils.HandleErr(err, "Couldn't move Deus page to Indexed")

	}

	deus.Index(wg)
}

func (deus *Deus) Sniff(wg *sync.WaitGroup) {
	log.Println("[+] Sniffing...")

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

		log.Println("[+] Wait 30s to continue sniff")
		time.Sleep(30 * time.Second)

	}
}

func (d *Deus) String() string { return "Deus" }
