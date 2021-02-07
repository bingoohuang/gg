package gtf

import (
	"fmt"
	htmlTemplate "html/template"
	"log"
	"math"
	"math/rand"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	textTemplate "text/template"
	"time"
)

var striptagsRegexp = regexp.MustCompile("<[^>]*?>")

// FindInSlice finds an element in the slice.
func FindInSlice(slice interface{}, f func(value interface{}) bool) int {
	var s reflect.Value

	if rv, ok := slice.(reflect.Value); ok {
		s = rv
	} else {
		s = reflect.ValueOf(slice)
	}

	if s.Kind() != reflect.Slice && s.Kind() != reflect.Array {
		return -1
	}

	for i := 0; i < s.Len(); i++ {
		if f(s.Index(i).Interface()) {
			return i
		}
	}

	return -1
}

var TextFuncMap = WrapRecover(textTemplate.FuncMap{
	"safeEq":   SafeEq,
	"contains": Contains,
	"replace": func(s1, s2 string) string {
		return strings.Replace(s2, s1, "", -1)
	},
	"findreplace": func(s1, s2, s3 string) string {
		return strings.Replace(s3, s1, s2, -1)
	},
	"title": func(s string) string {
		return strings.Title(s)
	},
	"default": func(arg, value interface{}) interface{} {
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
			if v.Len() == 0 {
				return arg
			}
		case reflect.Bool:
			if !v.Bool() {
				return arg
			}
		default:
			return value
		}

		return value
	},
	"length": func(value interface{}) int {
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return v.Len()
		case reflect.String:
			return len([]rune(v.String()))
		}

		return 0
	},
	"lower": func(s string) string {
		return strings.ToLower(s)
	},
	"upper": func(s string) string {
		return strings.ToUpper(s)
	},
	"truncatechars": func(n int, s string) string {
		if n < 0 {
			return s
		}

		r := []rune(s)
		rLength := len(r)

		if n >= rLength {
			return s
		}

		if n > 3 && rLength > 3 {
			return string(r[:n-3]) + "..."
		}

		return string(r[:n])
	},
	"urlencode": func(s string) string {
		return url.QueryEscape(s)
	},
	"wordcount": func(s string) int {
		return len(strings.Fields(s))
	},
	"divisibleby": func(arg interface{}, value interface{}) bool {
		var v float64
		switch value.(type) {
		case int, int8, int16, int32, int64:
			v = float64(reflect.ValueOf(value).Int())
		case uint, uint8, uint16, uint32, uint64:
			v = float64(reflect.ValueOf(value).Uint())
		case float32, float64:
			v = reflect.ValueOf(value).Float()
		default:
			return false
		}

		var a float64
		switch arg.(type) {
		case int, int8, int16, int32, int64:
			a = float64(reflect.ValueOf(arg).Int())
		case uint, uint8, uint16, uint32, uint64:
			a = float64(reflect.ValueOf(arg).Uint())
		case float32, float64:
			a = reflect.ValueOf(arg).Float()
		default:
			return false
		}

		return math.Mod(v, a) == 0
	},
	"lengthis": func(arg int, value interface{}) bool {
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return v.Len() == arg
		case reflect.String:
			return len([]rune(v.String())) == arg
		}

		return false
	},
	"trim": func(s string) string {
		return strings.TrimSpace(s)
	},
	"capfirst": func(s string) string {
		return strings.ToUpper(string(s[0])) + s[1:]
	},
	"pluralize": func(arg string, value interface{}) string {
		flag := false
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			flag = v.Int() == 1
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			flag = v.Uint() == 1
		default:
			return ""
		}

		if !strings.Contains(arg, ",") {
			arg = "," + arg
		}

		bits := strings.Split(arg, ",")

		if len(bits) > 2 {
			return ""
		}

		if flag {
			return bits[0]
		}

		return bits[1]
	},
	"yesno": func(yes, no string, value bool) string {
		if value {
			return yes
		}

		return no
	},
	"rjust": func(arg int, value string) string {
		n := arg - len([]rune(value))
		if n > 0 {
			value = strings.Repeat(" ", n) + value
		}

		return value
	},
	"ljust": func(arg int, value string) string {
		n := arg - len([]rune(value))

		if n > 0 {
			value = value + strings.Repeat(" ", n)
		}

		return value
	},
	"center": func(arg int, value string) string {
		n := arg - len([]rune(value))

		if n > 0 {
			left := n / 2
			right := n - left
			value = strings.Repeat(" ", left) + value + strings.Repeat(" ", right)
		}

		return value
	},
	"filesizeformat": func(value interface{}) string {
		var size float64

		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			size = float64(v.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			size = float64(v.Uint())
		case reflect.Float32, reflect.Float64:
			size = v.Float()
		default:
			return ""
		}

		var KB float64 = 1 << 10
		var MB float64 = 1 << 20
		var GB float64 = 1 << 30
		var TB float64 = 1 << 40
		var PB float64 = 1 << 50

		filesizeFormat := func(filesize float64, suffix string) string {
			return strings.Replace(fmt.Sprintf("%.1f %s", filesize, suffix), ".0", "", -1)
		}

		var result string
		if size < KB {
			result = filesizeFormat(size, "bytes")
		} else if size < MB {
			result = filesizeFormat(size/KB, "KB")
		} else if size < GB {
			result = filesizeFormat(size/MB, "MB")
		} else if size < TB {
			result = filesizeFormat(size/GB, "GB")
		} else if size < PB {
			result = filesizeFormat(size/TB, "TB")
		} else {
			result = filesizeFormat(size/PB, "PB")
		}

		return result
	},
	"apnumber": func(value interface{}) interface{} {
		name := [10]string{
			"one", "two", "three", "four", "five",
			"six", "seven", "eight", "nine",
		}

		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v.Int() < 10 {
				return name[v.Int()-1]
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if v.Uint() < 10 {
				return name[v.Uint()-1]
			}
		}

		return value
	},
	"intcomma": func(value interface{}) string {
		v := reflect.ValueOf(value)

		var x uint
		minus := false
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v.Int() < 0 {
				minus = true
				x = uint(-v.Int())
			} else {
				x = uint(v.Int())
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x = uint(v.Uint())
		default:
			return ""
		}

		var result string
		for x >= 1000 {
			result = fmt.Sprintf(",%03d%s", x%1000, result)
			x /= 1000
		}
		result = fmt.Sprintf("%d%s", x, result)

		if minus {
			result = "-" + result
		}

		return result
	},
	"ordinal": func(value interface{}) string {
		v := reflect.ValueOf(value)

		var x uint
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v.Int() < 0 {
				return ""
			}
			x = uint(v.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x = uint(v.Uint())
		default:
			return ""
		}

		suffixes := [10]string{"th", "st", "nd", "rd", "th", "th", "th", "th", "th", "th"}

		switch x % 100 {
		case 11, 12, 13:
			return fmt.Sprintf("%d%s", x, suffixes[0])
		}

		return fmt.Sprintf("%d%s", x, suffixes[x%10])
	},
	"first": func(value interface{}) interface{} {
		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.String:
			return string([]rune(v.String())[0])
		case reflect.Slice, reflect.Array:
			return v.Index(0).Interface()
		}

		return ""
	},
	"last": func(value interface{}) interface{} {
		switch v := reflect.ValueOf(value); v.Kind() {
		case reflect.String:
			str := []rune(v.String())
			return string(str[len(str)-1])
		case reflect.Slice, reflect.Array:
			return v.Index(v.Len() - 1).Interface()
		}

		return ""
	},
	"join": func(arg string, value []string) string {
		return strings.Join(value, arg)
	},
	"slice": func(start int, end int, value interface{}) interface{} {
		v := reflect.ValueOf(value)
		if start < 0 {
			start = 0
		}

		switch v.Kind() {
		case reflect.String:
			str := []rune(v.String())

			if end > len(str) {
				end = len(str)
			}

			return string(str[start:end])
		case reflect.Slice:
			return v.Slice(start, end).Interface()
		}
		return ""
	},
	"random": func(value interface{}) interface{} {
		rand.Seed(time.Now().UTC().UnixNano())
		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.String:
			str := []rune(v.String())
			return string(str[rand.Intn(len(str))])
		case reflect.Slice, reflect.Array:
			return v.Index(rand.Intn(v.Len())).Interface()
		}

		return ""
	},
	"randomintrange": func(min, max int, value interface{}) int {
		rand.Seed(time.Now().UTC().UnixNano())
		return rand.Intn(max-min) + min
	},
	"striptags": func(s string) string {
		return strings.TrimSpace(striptagsRegexp.ReplaceAllString(s, ""))
	},

	"panic": func(s interface{}) interface{} {
		panic(s)
	},
})

func WrapRecover(funcMap textTemplate.FuncMap) textTemplate.FuncMap {
	m := textTemplate.FuncMap{}
	for k, v := range funcMap {
		m[k] = recoverWrapFunc(v)
	}

	return m
}

func recoverWrapFunc(fn interface{}) interface{} {
	fnv := reflect.ValueOf(fn)
	return reflect.MakeFunc(fnv.Type(), func(args []reflect.Value) []reflect.Value {
		defer func() { // recovery will silently swallow all unexpected panics.
			if r := recover(); r != nil {
				log.Printf("E! recover from %v", r)
			}
		}()

		return fnv.Call(args)
	}).Interface()
}

// HtmlFuncMap defines the html template functions map.
var HtmlFuncMap = htmlTemplate.FuncMap(TextFuncMap)

// NewHtmlTemplate is a wrapper function of template.New(https://golang.org/pkg/html/template/#New).
// It automatically adds the gtf functions to the template's function map
// and returns template.Template(http://golang.org/pkg/html/template/#Template).
func NewHtmlTemplate(name string) *htmlTemplate.Template {
	return htmlTemplate.New(name).Funcs(HtmlFuncMap)
}

// NewTextTemplate is a wrapper function of template.New(https://golang.org/pkg/text/template/#New).
// It automatically adds the gtf functions to the template's function map
// and returns template.Template(http://golang.org/pkg/text/template/#Template).
func NewTextTemplate(name string) *textTemplate.Template {
	return textTemplate.New(name).Funcs(TextFuncMap)
}

// Inject injects gtf functions into the passed FuncMap.
// It does not overwrite the original function which have same name as a gtf function.
func Inject(funcs map[string]interface{}, force bool, prefix string) {
	for k, v := range TextFuncMap {
		if force {
			funcs[prefix+k] = v
		} else if _, ok := funcs[k]; !ok {
			funcs[prefix+k] = v
		}
	}
}

func SafeEq(v interface{}, name string, defaultValue, compareValue interface{}) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Map:
		if v := rv.MapIndex(reflect.ValueOf(name)); v.IsValid() {
			return v.Interface() == compareValue
		}
	case reflect.Struct:
		if v := rv.FieldByName(name); v.IsValid() {
			return v.Interface() == compareValue
		}
	}

	return defaultValue == compareValue
}

func Contains(v interface{}, name string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Map:
		return rv.MapIndex(reflect.ValueOf(name)).IsValid()
	case reflect.Struct:
		return rv.FieldByName(name).IsValid()
	case reflect.Slice, reflect.Array:
		f := func(v interface{}) bool { return v == name }
		return FindInSlice(rv, f) >= 0
	}

	return false
}
