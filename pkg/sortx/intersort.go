package sortx

import (
	"fmt"
	"reflect"
	"sort"
)

func less(x, y interface{}) bool {
	justX := fmt.Sprint(map[interface{}]struct{}{x: {}})
	justY := fmt.Sprint(map[interface{}]struct{}{y: {}})
	return fmt.Sprint(map[interface{}]struct{}{
		x: {},
		y: {},
	}) == fmt.Sprintf("map[%v %v]", justX[4:len(justX)-1], justY[4:len(justY)-1])
}

// Slice implements sort.Interface for arbitrary objects, according to the map ordering of the fmt package.
type Slice []interface{}

func (is Slice) Len() int           { return len(is) }
func (is Slice) Swap(i, j int)      { is[i], is[j] = is[j], is[i] }
func (is Slice) Less(i, j int) bool { return less(is[i], is[j]) }

type (
	Lesser     func(a, b interface{}) bool
	SortConfig struct {
		Less Lesser
	}
)

type LessSlice struct {
	Slice
	Lesser
}

func WrapLesser(s Slice, lesser Lesser) LessSlice {
	return LessSlice{Slice: s, Lesser: lesser}
}

func (is LessSlice) Less(i, j int) bool { return is.Lesser(is.Slice[i], is.Slice[j]) }

type SortOption func(*SortConfig)

func WithLess(f Lesser) SortOption { return func(c *SortConfig) { c.Less = f } }

// Sort sorts arbitrary objects according to the map ordering of the fmt package.
func Sort(slice interface{}, options ...SortOption) {
	config := &SortConfig{}
	for _, option := range options {
		option(config)
	}
	if config.Less == nil {
		config.Less = less
	}

	val := reflect.ValueOf(slice)
	if val.Type().Kind() != reflect.Slice {
		panic("sortx: cannot sort non-slice type")
	}
	sort.Slice(slice, func(i, j int) bool {
		return config.Less(val.Index(i).Interface(), val.Index(j).Interface())
	})
}
