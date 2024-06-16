package ishtari

import (
	"fmt"
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

type Ishtari struct {
	db *database.Database
}

func NewIshtari(db *database.Database) *Ishtari {
	return &Ishtari{
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

		// E.g. https://ishtari.com.gh/USB-Desktop-Microphone-With-Tripod-/p=815
		productLink := fmt.Sprintf("https://ishtari.com.gh%s", link.Attrs()["href"])

		canQueue, err := db.CanQueueUrl(productLink)
		if utils.HandleErr(err, "Failed to get Ishtari queue") {
			return
		}

		if canQueue {
			err = db.AddToQueue(data.UrlQueue{
				URL:    productLink,
				Source: "Ishtari",
			})

			utils.HandleErr(err, "Failed to add Ishtari to queue")
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
func extractProducts(href string) ([]soup.Root, int) {
	log.Println("[+] Extracting products from", href)

	resp := utils.FetchPage(href, "rod")

	doc := soup.HTMLParse(resp)

	paginationEl := doc.Find("ul", "class", "category-pagination")

	totalPages := 0

	var err error

	if paginationEl.Error == nil {
		paginationChildren := paginationEl.Children()

		totalPages, err = strconv.Atoi(paginationChildren[len(paginationChildren)-2].FullText())
		if err != nil {
			log.Fatalln(err)
		}

	}

	return doc.FindAll("a", "class", "false"), totalPages
}

func (ishtari *Ishtari) Index(wg *sync.WaitGroup) {
	log.Println("[+] Indexing Ishtari...")

	pages, err := ishtari.db.GetCrawledPages("Ishtari")
	if utils.HandleErr(err, "Failed to get Ishtari crawled pages") {
		return
	}

	if len(pages) == 0 {
		log.Println("[+] No pages to index for Ishtari!")
		log.Println("[+] Waiting 60s to continue indexing...")

		time.Sleep(60 * time.Second)

		ishtari.Index(wg)

		wg.Done()
		return
	}

	for _, page := range pages {
		parsedPage := soup.HTMLParse(page.HTML)

		productNameEl := parsedPage.Find("h1", "class", "text-d22")

		if productNameEl.Error != nil {
			err = ishtari.db.DeleteCrawledPage(page.URL)
			utils.HandleErr(err, "Failed to delete Ishtari crawled page")
			continue
		}

		productName := productNameEl.Text()

		productPriceStirng := strings.ReplaceAll(parsedPage.Find("span", "class", "false").Text(), " GH¢", "")

		price, err := strconv.ParseFloat(strings.ReplaceAll(productPriceStirng, ",", ""), 64)
		if utils.HandleErr(err, "Failed to parse Ishtari product price") {
			return
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

		err = ishtari.db.IndexProduct(productData)
		if utils.HandleErr(err, "Failed to index Ishtari product") {
			return
		}

		err = ishtari.db.MovePageToIndexed(page)
		utils.HandleErr(err, "Failed to move Ishtari page to indexed")
	}

	ishtari.Index(wg)
}

func (ishtari *Ishtari) Sniff(wg *sync.WaitGroup) {
	log.Println("[+] Sniffing...")

	defer wg.Done()

	html := utils.FetchPage("https://ishtari.com.gh/", "rod")

	doc := soup.HTMLParse(html)

	links := doc.FindAll("a", "class", "text-d13")

	utils.ShuffleLinks(links)

	for _, link := range links {
		// E.g. /Electronics/c=1023
		categoryLink := link.Attrs()["href"]

		categoryLink = fmt.Sprintf("https://ishtari.com.gh%s", categoryLink)

		products, totalPages := extractProducts(categoryLink)

		queueProducts(ishtari.db, products)

		for i := 2; i <= totalPages; i++ {
			go func(i int) {
				// E.g. https://ishtari.com.gh/Back-To-School/c=918?page=6
				pageLink := fmt.Sprintf("%s?page=%d", categoryLink, i)

				pageProducts, _ := extractProducts(pageLink)

				queueProducts(ishtari.db, pageProducts)
			}(i)
		}

		log.Println("[+] Wait 120s to continue sniff")
		time.Sleep(120 * time.Second)

	}
}

func (ishtari *Ishtari) String() string { return "Ishtari" }
