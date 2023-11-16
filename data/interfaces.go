package data

import "sync"

type Sniffer interface {
	Sniff(wg *sync.WaitGroup)
}

type Indexer interface {
	Index(wg *sync.WaitGroup)
}
