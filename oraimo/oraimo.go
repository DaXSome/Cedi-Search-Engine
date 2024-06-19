package oraimo

import (
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
	for _, link := range products {
		// E.g. https://gh.oraimo.com/oraimo-freepods-lite-40-hour-playtime-enc-true-wireless-earbuds.html
		productLink := link.Attrs()["href"]

		canQueue, err := db.CanQueueUrl(productLink)
		if utils.HandleErr(err, "Failed to get Oraimo queue") {
			return
		}

		if canQueue {
			err = db.AddToQueue(data.UrlQueue{
				URL:    productLink,
				Source: "Oraimo",
			})

			utils.HandleErr(err, "Failed to add Oraimo to queue")
		} else {
			utils.Logger("sniffer", "[+] Skipping", productLink)
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
	utils.Logger("sniffer", "[+] Extracting products from ", href)

	resp := utils.FetchPage(href, "rod")

	doc := soup.HTMLParse(resp)

	return doc.FindAll("a", "class", "product")
}

func (oraimo *Oraimo) Index(wg *sync.WaitGroup) {
	utils.Logger("indexer", "[+] Indexing Oraimo...")

	pages, err := oraimo.db.GetCrawledPages("Oraimo")
	if utils.HandleErr(err, "Failed to get Oraimo crawled pages") {
		return
	}

	if len(pages) == 0 {
		utils.Logger("indexer", "[+] No pages to index for Oraimo!")
		utils.Logger("indexer", "[+] Waiting 60s to continue indexing...")

		time.Sleep(60 * time.Second)

		oraimo.Index(wg)

		wg.Done()
		return
	}

	for _, page := range pages {
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

		err = oraimo.db.MovePageToIndexed(page)
		utils.HandleErr(err, "Failed to move Oraimo page to indexed")
	}

	oraimo.Index(wg)
}

func (oraimo *Oraimo) Sniff(wg *sync.WaitGroup) {
	utils.Logger("sniffer", "[+] Sniffing...")

	defer wg.Done()

	resp := utils.FetchPage("https://gh.oraimo.com/", "rod")

	doc := soup.HTMLParse(resp)

	links := doc.FindAll("a", "role", "menuitem")

	utils.ShuffleLinks(links)

	for _, link := range links {
		// E.g. https://gh.oraimo.com/products/lifestyle/electric-toothbrush.html
		categoryLink := link.Attrs()["href"]

		if strings.Contains(categoryLink, "products") {

			products := extractProducts(categoryLink)

			queueProducts(oraimo.db, products)

			utils.Logger("sniffer", "[+] Wait 15s to continue sniff")
			time.Sleep(15 * time.Second)
		}

	}
}

func (oraimo *Oraimo) String() string { return "Oraimo" }
