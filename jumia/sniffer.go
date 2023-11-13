package jumia

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/models"

	"github.com/anaskhan96/soup"
)

func queueProducts(db *database.Database, products []soup.Root) {
	for _, link := range products {
		// E.g. https://www.jumia.com.gh/jameson-irish-whiskey-750ml-51665215.html
		productLink := fmt.Sprintf("https://www.jumia.com.gh%s", link.Attrs()["href"])

		if db.CanQueueUrl(productLink) {
			db.AddToQueue(models.UrlQueue{
				URL:    productLink,
				Source: "Jumia",
			})
		} else {
			log.Println("[+] Skipping", productLink)
		}

	}
}

func extractProducts(href string) ([]soup.Root, int) {
	log.Println("[+] Extracting products from", href)
	resp, err := soup.Get(href)

	if err != nil {
		log.Fatalln(err)
	}

	doc := soup.HTMLParse(resp)

	totalPagesEl := doc.FindAll("a", "class", "pg")

	// E.g. /groceries/?page=50#catalog-listing
	lastPageLink := totalPagesEl[len(totalPagesEl)-1].Attrs()["href"]

	totalPages, err := strconv.Atoi(strings.Split(strings.Split(lastPageLink, "=")[1], "#")[0])

	if err != nil {
		log.Println(err)
	}

	log.Println("[+] Wait 60s to continue extracting products")
	time.Sleep(60 * time.Second)

	return doc.FindAll("a", "class", "core"), totalPages
}

type Sniffer interface {
	Sniff(wg *sync.WaitGroup)
}

type SnifferImpl struct {
	db *database.Database
}

func NewSniffer(database *database.Database) *SnifferImpl {
	return &SnifferImpl{
		db: database,
	}
}

func (sl *SnifferImpl) Sniff(wg *sync.WaitGroup) {
	log.Println("[+] Sniffing...")

	defer wg.Done()

	soup.Header("User-Agent", "cedisearchbot/0.1 (+https://cedi-search.vercel.app/about)")

	resp, err := soup.Get("https://www.jumia.com.gh")

	if err != nil {
		log.Fatalln(err)
	}

	doc := soup.HTMLParse(resp)

	links := doc.FindAll("a", "role", "menuitem")

	for _, link := range links {
		// E.g. https://www.jumia.com.gh/groceries
		categoryLink := link.Attrs()["href"]

		if categoryLink != "" {
			if !strings.Contains(categoryLink, "jumia") {
				categoryLink = fmt.Sprintf("https://www.jumia.com.gh%s", categoryLink)
			}

			products, totalPages := extractProducts(categoryLink)

			queueProducts(sl.db, products)

			for i := 2; i <= totalPages; i++ {
				// E.g. https://www.jumia.com.gh/groceries?page=2
				pageLink := fmt.Sprintf("%s?page=%d", categoryLink, i)

				pageProducts, _ := extractProducts(pageLink)

				queueProducts(sl.db, pageProducts)

			}

			log.Println("[+] Wait 60s to continue sniff")
			time.Sleep(60 * time.Second)

		}
	}
}
