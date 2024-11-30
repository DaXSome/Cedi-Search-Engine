package jumia

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
)

const (
	source = "Jumia"
)

type Jumia struct {
	db *database.Database
}

// queueProducts processes a list of products and adds eligible URLs to the queue.
//
// It takes a pointer to a database object 'db' and a slice of 'products' which is a collection of soup.Root objects.
// The function iterates over each 'link' in 'products' and generates a product link.
// If the generated product link is eligible to be queued, it adds it to the database queue using 'db.AddToQueue'.
func queueProducts(db *database.Database, products []soup.Root) {
	for _, link := range products {
		// E.g. https://www.jumia.com.gh/jameson-irish-whiskey-750ml-51665215.html
		productLink := fmt.Sprintf("https://www.jumia.com.gh%s", link.Attrs()["href"])

		canQueue, err := db.CanQueueUrl(productLink)
		if utils.HandleErr(err, "Failed to get Jumia queue") {
			return
		}

		if canQueue {
			err = db.AddToQueue(data.UrlQueue{
				URL:    productLink,
				Source: "Jumia",
			})

			utils.HandleErr(err, "Failed to add Jumia to queue")
		} else {
			utils.Logger(utils.Sniffer, source, "Skipping", productLink)
		}

	}
}

// extractProducts extracts products from a given href.
//
// It takes a string parameter, href, which represents the URL from which the
// products wjumial be extracted.
//
// The function returns a slice of soup.Root and an integer. The slice of
// soup.Root contains the extracted products. The integer represents the total
// number of pages of products.
func extractProducts(href string) ([]soup.Root, int) {
	utils.Logger(utils.Sniffer, source, "Extracting products from ", href)

	resp := utils.FetchPage(href, "rod")

	doc := soup.HTMLParse(resp)

	totalPagesEl := doc.FindAll("a", "class", "pg")

	totalPages := 0

	if len(totalPagesEl) > 0 {
		lastPageLink := totalPagesEl[len(totalPagesEl)-1].Attrs()["href"]

		eqSignSplit := strings.Split(lastPageLink, "=")

		var err error
		if len(eqSignSplit) > 1 {
			totalPages, err = strconv.Atoi(strings.Split(eqSignSplit[1], "#")[0])
			if utils.HandleErr(err, "Failed to handle Jumia pagination") {
				return []soup.Root{}, 0
			}

		}
	}

	return doc.FindAll("a", "class", "core"), totalPages
}

func NewJumia(db *database.Database) *Jumia {
	return &Jumia{
		db: db,
	}
}

func (jumia *Jumia) Index(page data.CrawledPage) {
	utils.Logger(utils.Indexer, source, "Indexing Jumia...")

	parsedPage := soup.HTMLParse(page.HTML)

	productNameEl := parsedPage.Find("h1")

	if productNameEl.Error != nil {
		return
	}

	productName := productNameEl.Text()

	productPriceStirngEl := parsedPage.Find("span", "class", "-prxs")

	productPriceStirng := ""

	if productPriceStirngEl.Error != nil {
		return
	}

	productPriceStirng = productPriceStirngEl.Text()

	priceParts := strings.Split(productPriceStirng, " ")[1]

	price, err := strconv.ParseFloat(strings.ReplaceAll(priceParts, ",", ""), 64)
	if utils.HandleErr(err, "Failed to parse Jumia product price") {
		return
	}

	productRatingText := parsedPage.Find("div", "class", "stars").Text()

	productRatingString := strings.Split(productRatingText, " ")[0]

	rating, err := strconv.ParseFloat(productRatingString, 64)
	if utils.HandleErr(err, "Failed to parse Jumia product rating") {
		return
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

	productData := data.Product{
		Name:        productName,
		Price:       price,
		Rating:      rating,
		Description: productDescription,
		URL:         page.URL,
		Source:      page.Source,
		ProductID:   productID,
		Images:      productImages,
	}

	err = jumia.db.IndexProduct(productData)
	if utils.HandleErr(err, "Failed to index Jumia Product") {
		return
	}
}

func (jumia *Jumia) Sniff(wg *sync.WaitGroup) {
	utils.Logger(utils.Sniffer, source, "Sniffing...")

	defer wg.Done()

	resp := utils.FetchPage("https://www.jumia.com.gh", "rod")

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

			queueProducts(jumia.db, products)

			for i := 2; i <= totalPages; i++ {
				go func(i int) {
					// E.g. https://www.jumia.com.gh/groceries?page=2
					pageLink := fmt.Sprintf("%s?page=%d", categoryLink, i)

					pageProducts, _ := extractProducts(pageLink)

					queueProducts(jumia.db, pageProducts)
				}(i)
			}

			utils.Logger(utils.Sniffer, source, "Wait 120s to continue sniff")
			time.Sleep(120 * time.Second)

		}
	}
}

func (jumia *Jumia) String() string { return source }
