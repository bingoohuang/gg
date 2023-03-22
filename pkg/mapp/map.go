package mapp

import (
	"reflect"
	"sort"

	"github.com/bingoohuang/gg/pkg/sortx"
)

// Keys 返回Map的key切片
func Keys(m interface{}) []string {
	return KeysX(m).([]string)
}

// KeysSorted 返回Map排序后的key切片
func KeysSorted(m interface{}) []string {
	keys := Keys(m)
	sort.Strings(keys)
	return keys
}

// KeysX 返回Map的key切片
func KeysX(m interface{}) interface{} {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil
	}

	keyType := v.Type().Key()
	ks := reflect.MakeSlice(reflect.SliceOf(keyType), v.Len(), v.Len())

	for i, key := range v.MapKeys() {
		ks.Index(i).Set(key)
	}

	return ks.Interface()
}

// KeysSortedX 返回Map排序后的key切片
func KeysSortedX(m interface{}) interface{} {
	mv := reflect.ValueOf(m)
	if mv.Kind() != reflect.Map {
		return nil
	}

	mapLen := mv.Len()
	ks := reflect.MakeSlice(reflect.SliceOf(mv.Type().Key()), mapLen, mapLen)
	for i, k := range mv.MapKeys() {
		ks.Index(i).Set(k)
	}

	ksi := ks.Interface()
	sortx.Sort(ksi)
	return ksi
}

// Values 返回Map的value切片
func Values(m interface{}) []string {
	return MapValuesX(m).([]string)
}

// MapValuesX 返回Map的value切片
func MapValuesX(m interface{}) interface{} {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil
	}

	sl := reflect.MakeSlice(reflect.SliceOf(v.Type().Elem()), v.Len(), v.Len())
	for i, key := range v.MapKeys() {
		sl.Index(i).Set(v.MapIndex(key))
	}

	return sl.Interface()
}

// GetOr get value from m by key or returns the defaultValue
func GetOr(m, k, defaultValue interface{}) interface{} {
	mv := reflect.ValueOf(m)
	if mv.Kind() != reflect.Map {
		return nil
	}

	v := mv.MapIndex(reflect.ValueOf(k))
	if v.IsValid() {
		return v.Interface()
	}

	return defaultValue
}

// WalkMap iterates the map by iterFunc.
func WalkMap(m interface{}, iterFunc interface{}) {
	mv := reflect.ValueOf(m)
	if mv.Kind() != reflect.Map {
		return
	}

	mapLen := mv.Len()
	ks := reflect.MakeSlice(reflect.SliceOf(mv.Type().Key()), mapLen, mapLen)

	for i, k := range mv.MapKeys() {
		ks.Index(i).Set(k)
	}

	ksi := ks.Interface()
	sortx.Sort(ksi)

	funcValue := reflect.ValueOf(iterFunc)

	for j := 0; j < mapLen; j++ {
		k := ks.Index(j)
		v := mv.MapIndex(k)
		funcValue.Call([]reflect.Value{k, v})
	}
}

// Clone clones a map[string]string.
func Clone(m map[string]string) map[string]string {
	c := make(map[string]string)
	for k, v := range m {
		c[k] = v
	}
	return c
}
