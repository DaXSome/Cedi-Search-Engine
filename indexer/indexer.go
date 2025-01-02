package indexer

import (
	"fmt"

	"github.com/Cedi-Search/Cedi-Search-Engine/data"
	"github.com/Cedi-Search/Cedi-Search-Engine/database"
	"github.com/Cedi-Search/Cedi-Search-Engine/utils"
	"github.com/anaskhan96/soup"
)

type Indexer struct {
	db *database.Database
}

func NewIndexer(database *database.Database) *Indexer {
	return &Indexer{
		db: database,
	}
}

func (indexer *Indexer) Index(page data.CrawledPage) error {
	utils.Logger(utils.Indexer, page.Source, fmt.Sprintf("Indexing %v...", page.Source))

	parsedPage := soup.HTMLParse(page.HTML)

	productData := make(map[string]interface{})

	productData["url"] = page.URL
	productData["source"] = page.Source

	for _, attrib := range page.Attribs {
		args := []string{attrib.Selector.Element}

		if attrib.Selector.Attribute != "" && attrib.Selector.Attribute != "" {
			args = append(args, attrib.Selector.Attribute, attrib.Selector.Value)
		}

		if attrib.IsArray {
			els := parsedPage.FindAll(args...)

			elItems := []string{}

			for _, el := range els {

				if el.Error != nil {
					continue
				}

				var item string

				if attrib.ChildAttrib != "" {
					item = el.Attrs()[attrib.ChildAttrib]
				} else {
					item = el.FullText()
				}

				elItems = append(elItems, item)

			}

			productData[attrib.Label] = elItems

		} else {

			el := parsedPage.Find(args...)

			if el.Error != nil {
				continue
			}

			var item string

			if attrib.ChildAttrib != "" {
				item = el.Attrs()[attrib.ChildAttrib]
			}

			item = el.FullText()

			productData[attrib.Label] = item

		}

		err := indexer.db.IndexProduct(productData)
		if err != nil {
			return err
		}
	}

	return nil
}
