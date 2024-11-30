package oraimo

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"
	"github.com/anaskhan96/soup"
	"github.com/google/uuid"
)

const (
	source = "Oraimo"
)

type Oraimo struct {
	db *database.Database
}

func NewOraimo(db *database.Database) *Oraimo {
	return &Oraimo{
		db: db,
	}
}

// queueProducts processes a list of products and adds eligible URLs to the queue.
//
// It takes a pointer to a database object 'db' and a slice of 'products' which is a collection of soup.Root objects.
// The function iterates over each 'link' in 'products' and generates a product link.
// If the generated product link is eligible to be queued, it adds it to the database queue using 'db.AddToQueue'.
func queueProducts(db *database.Database, products []soup.Root) {
	for _, product := range products {

		productMetaTag := product.Find("a", "class", "product-img")

		// E.g./product/oraimo-boompop-2-powerful-deep-bass-dual-device-connectivity-wireless-headset?ean=4894947008030
		productLink := productMetaTag.Attrs()["href"]

		fmtedProductLink := fmt.Sprintf("https://gh.oraimo.com%s", strings.Split(productLink, "?")[0])

		canQueue, err := db.CanQueueUrl(fmtedProductLink)
		if utils.HandleErr(err, "Failed to get Oraimo queue") {
			return
		}

		if canQueue {
			err = db.AddToQueue(data.UrlQueue{
				URL:    fmtedProductLink,
				Source: "Oraimo",
			})

			utils.HandleErr(err, "Failed to add Oraimo to queue")
		} else {
			utils.Logger(utils.Sniffer, source, "Skipping", fmtedProductLink)
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
	utils.Logger(utils.Sniffer, source, "Extracting products from ", href)

	resp := utils.FetchPage(href, "rod")

	doc := soup.HTMLParse(resp)

	return doc.FindAll("div", "class", "site-product")
}

func (oraimo *Oraimo) Index(page data.CrawledPage) {
	utils.Logger(utils.Indexer, source, "Indexing Oraimo...")

	parsedPage := soup.HTMLParse(page.HTML)

	productName := parsedPage.Find("h1").FullText()
	productName = strings.ReplaceAll(productName, "\n", "")
	productName = strings.Trim(productName, " ")

	productPriceStirng := parsedPage.Find("span", "class", "price").Text()
	productPriceStirng = strings.ReplaceAll(productPriceStirng, "₵", "")

	price, err := strconv.ParseFloat(strings.ReplaceAll(productPriceStirng, ",", ""), 64)
	if utils.HandleErr(err, "Failed to parse Oraimo product price") {
		return
	}

	rating := 0.0

	ratingEl := parsedPage.Find("div", "class", "rating-result")

	if ratingEl.Error == nil {

		productRatingText := ratingEl.Attrs()["title"]

		rating, err = strconv.ParseFloat(productRatingText, 64)
		if utils.HandleErr(err, "Failed to convert Oraimo product price") {
			return
		}
	}

	productDescription := parsedPage.Find("div", "id", "description").FullText()

	productID := uuid.New()

	productImagesEl := parsedPage.FindAll("img", "class", "fotorama__img")

	productImages := []string{}

	for _, el := range productImagesEl {
		productImages = append(productImages, el.Attrs()["src"])
	}

	productData := data.Product{
		Name:        productName,
		Price:       price,
		Rating:      rating,
		Description: productDescription,
		URL:         page.URL,
		Source:      page.Source,
		ProductID:   productID.String(),
		Images:      productImages,
	}

	err = oraimo.db.IndexProduct(productData)
	if utils.HandleErr(err, "Failed to index Oraimo product") {
		return
	}
}

func (oraimo *Oraimo) Sniff(wg *sync.WaitGroup) {
	utils.Logger(utils.Sniffer, source, "Sniffing...")

	defer wg.Done()

	links := []string{
		"https://gh.oraimo.com/oraimo-daily-deals.html",
		"https://gh.oraimo.com/promotion/free-gifts",
		"https://gh.oraimo.com/collections/audio",
		"https://gh.oraimo.com/collections/power",
		"https://gh.oraimo.com/collections/smart-and-office",
		"https://gh.oraimo.com/collections/personal-care",
		"https://gh.oraimo.com/collections/home-appliances",
	}

	utils.ShuffleLinks(links)

	for _, link := range links {
		// E.g. https://gh.oraimo.com/products/lifestyle/electric-toothbrush.html

		products := extractProducts(link)

		queueProducts(oraimo.db, products)

		utils.Logger(utils.Sniffer, source, "Wait 15s to continue sniff")
		time.Sleep(15 * time.Second)
	}
}

func (oraimo *Oraimo) String() string { return source }
