package data

import (
	"sync"
)

type Target interface {
	Sniff(wg *sync.WaitGroup)
	Index(wg *sync.WaitGroup)
	String() string
}
