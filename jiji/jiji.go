package jiji

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"
	"github.com/anaskhan96/soup"
)

type Jiji struct {
	db *database.Database
}

func NewJiji(db *database.Database) *Jiji {
	return &Jiji{
		db: db,
	}
}

// ShuffleLinks shuffles the order of links.
func shuffleLinks(links []string) {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	rand.Shuffle(len(links), func(i, j int) {
		links[i], links[j] = links[j], links[i]
	})
}

// queueProducts processes a list of products and adds eligible URLs to the queue.
//
// It takes a pointer to a database object 'db' and a slice of 'products' which is a collection of soup.Root objects.
// The function iterates over each 'link' in 'products' and generates a product link.
// If the generated product link is eligible to be queued, it adds it to the database queue using 'db.AddToQueue'.
func queueProducts(db *database.Database, products []soup.Root) {
	for _, link := range products {
		// E.g. https://jiji.com.gh/us-embassy-area/commercial-properties/apartments-yZ4tX1iUJB0rSdhAdhf1UA7x.html?page=2&pos=1&cur_pos=1&ads_per_page=23&ads_count=63809&lid=Fmd1TGLFlcaLNkMG&indexPosition=0
		productLink := fmt.Sprintf("https://jiji.com.gh%s", link.Attrs()["href"])
		productLink = strings.Split(productLink, "?")[0]

		canQueue, err := db.CanQueueUrl(productLink)
		if utils.HandleErr(err, "Can't get queue for Jiji ") {
			return
		}

		if canQueue {
			err = db.AddToQueue(data.UrlQueue{
				URL:    productLink,
				Source: "Jiji",
			})

			utils.HandleErr(err, "Failed to add Jiji url to queue")
		} else {
			utils.Logger("sniffing", "[+] Skipping", productLink)
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

	return doc.FindAll("a", "class", "b-list-advert-base")
}

func (jiji *Jiji) Index(page data.CrawledPage) {
	utils.Logger("indexer", "[+] Indexing Jiji...")

	parsedPage := soup.HTMLParse(page.HTML)

	// E.g Kia Sorento 2.5 D Automatic 2003 Red in Akuapim South - Cars, Gabriel Sokah | Jiji.com.gh
	productNameEl := parsedPage.Find("title")

	if productNameEl.Error != nil {
		return
	}

	productName := productNameEl.Text()
	productName = strings.Split(productName, " in ")[0]

	productPriceEl := parsedPage.Find("div", "itemprop", "price")

	if productPriceEl.Error != nil {
		return
	}

	productPriceString := productPriceEl.Attrs()["content"]

	if productPriceString == "" {
		return
	}

	price, err := strconv.ParseFloat(productPriceString, 64)
	if utils.HandleErr(err, "Failed to convert Jiji product price") {
		return
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
		imageEl := parsedPage.Find("img", "class", "b-slider-image")

		if imageEl.Error == nil {
			productImages = append(productImages, imageEl.Attrs()["src"])
		}
	}

	productData := data.Product{
		Name:        productName,
		Price:       price,
		Rating:      0,
		Description: productDescription,
		URL:         page.URL,
		Source:      page.Source,
		ProductID:   productID,
		Images:      productImages,
	}

	err = jiji.db.IndexProduct(productData)
	if utils.HandleErr(err, "Failed to index Jiji product") {
		return
	}
}

func (jiji *Jiji) Sniff(wg *sync.WaitGroup) {
	utils.Logger("sniffer", "[+] Sniffing...")

	defer wg.Done()

	categories := []string{
		"vehicles",
		"real-estate",
		"mobile-phones-tablets",
		"electronics",
		"home-garden",
		"health-and-beauty",
		"fashion-and-beauty",
		"hobbies-art-sport",
		"seeking-work-cvs",
		"services",
		"jobs",
		"babies-and-kids",
		"animals-and-pets",
		"agriculture-and-foodstuff",
		"office-and-commercial-equipment-tools",
		"repair-and-construction",
	}

	shuffleLinks(categories)

	for _, category := range categories {

		categoryLink := fmt.Sprintf("https://jiji.com.gh/%s", category)

		for i := 1; i <= 1000; i++ {
			pageLink := fmt.Sprintf("%s?page=%d", categoryLink, i)

			// E.g. https://jiji.com.gh/repair-and-construction?page=992
			pageProducts := extractProducts(pageLink)

			queueProducts(jiji.db, pageProducts)

			if i%50 == 0 {
				utils.Logger("sniffing", "[+] Wait 120s to continue sniff")
				time.Sleep(120 * time.Second)
			}

		}

	}
}

func (jiji *Jiji) String() string { return "Jiji" }
