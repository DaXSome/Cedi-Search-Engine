package data

import (
	"sync"
)

type Target interface {
	Sniff(wg *sync.WaitGroup)
	Index(page CrawledPage)
	String() string
}
