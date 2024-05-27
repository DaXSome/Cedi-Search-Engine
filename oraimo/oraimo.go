package oraimo

import (
	"log"
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

		if db.CanQueueUrl(productLink) {
			db.AddToQueue(data.UrlQueue{
				URL:    productLink,
				Source: "Oraimo",
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
func extractProducts(href string) []soup.Root {
	log.Println("[+] Extracting products from", href)

	resp := utils.FetchPage(href, "rod")

	doc := soup.HTMLParse(resp)

	return doc.FindAll("a", "class", "product")
}

func (oraimo *Oraimo) Index(wg *sync.WaitGroup) {
	log.Println("[+] Indexing Oraimo...")

	pages := oraimo.db.GetCrawledPages("Oraimo")

	if len(pages) == 0 {
		log.Println("[+] No pages to index for Oraimo!")
		log.Println("[+] Waiting 60s to continue indexing...")

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
		if err != nil {
			log.Fatalln(err)
		}

		rating := 0.0

		ratingEl := parsedPage.Find("div", "class", "rating-result")

		if ratingEl.Error == nil {

			productRatingText := ratingEl.Attrs()["title"]

			rating, err = strconv.ParseFloat(productRatingText, 64)
			if err != nil {
				log.Fatalln(err)
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

		oraimo.db.IndexProduct(productData)
		oraimo.db.MovePageToIndexed(page)

	}

	oraimo.Index(wg)
}

func (oraimo *Oraimo) Sniff(wg *sync.WaitGroup) {
	log.Println("[+] Sniffing...")

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

			log.Println("[+] Wait 15s to continue sniff")
			time.Sleep(15 * time.Second)
		}

	}
}

func (oraimo *Oraimo) String() string { return "Oraimo" }
