package data

import (
	"sync"
)

type T interface {
	Sniff(wg *sync.WaitGroup)
	Index(page CrawledPage)
	String() string
}
