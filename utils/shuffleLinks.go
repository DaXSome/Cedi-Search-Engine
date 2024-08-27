package utils

import (
	"math/rand"
	"time"

	"github.com/anaskhan96/soup"
)

type Link interface {
	soup.Root | string
}

// ShuffleLinks shuffles the order of links.
func ShuffleLinks[T Link](links []T) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(links), func(i, j int) {
		links[i], links[j] = links[j], links[i]
	})
}
