package randx

import (
	"math/rand"
	"time"
)

// Shuffle pseudo-randomizes the order of elements using the default Source.
func Shuffle(a []string) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	return a
}
