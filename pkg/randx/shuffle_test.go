package randx

import (
	"fmt"
	"testing"
)

func TestShuffle(t *testing.T) {
	a := []string{"a", "b", "c"}
	Shuffle(a)
	fmt.Println(a)
	fmt.Println(ShuffleSs(a))
}
