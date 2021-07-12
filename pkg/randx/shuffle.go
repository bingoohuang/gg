package randx

import (
	"reflect"
)

// Shuffle pseudo-randomizes the order of elements using the default Source.
// https://stackoverflow.com/questions/12264789/shuffle-array-in-go
func Shuffle(slice interface{}) {
	rv := reflect.ValueOf(slice)
	swap := reflect.Swapper(slice)
	for i := rv.Len() - 1; i > 0; i-- {
		j := IntN(i + 1)
		swap(i, j)
	}
}

func CopySlice(s interface{}) interface{} {
	t, v := reflect.TypeOf(s), reflect.ValueOf(s)
	c := reflect.MakeSlice(t, v.Len(), v.Len())
	reflect.Copy(c, v)
	return c.Interface()
}

func ShuffleSs(a []string) []string {
	b := append([]string(nil), a...)
	swap := func(i, j int) { b[i], b[j] = b[j], b[i] }
	for i := len(b) - 1; i > 0; i-- {
		swap(i, IntN(i+1))
	}
	return b
}
