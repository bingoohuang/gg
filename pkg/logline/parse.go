package logline

import (
	"errors"
	"github.com/bingoohuang/gg/pkg/ss"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var filters = map[string]Converter{
	"path": UriPath(),
}

type Pattern struct {
	Pattern    string
	Converters map[string]Converter
	Dots       []Dot
}

// SliceToString preferred for large body payload (zero allocation and faster)
func SliceToString(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }

func (p Pattern) Names() (names []string) {
	for _, dot := range p.Dots {
		if dot.Valid() {
			names = append(names, dot.Name)
		}
	}
	return
}

func (p Pattern) ParseBytes(s []byte) map[string]interface{} {
	return p.Parse(SliceToString(s))
}

func (p Pattern) Parse(s string) map[string]interface{} {
	m := make(map[string]interface{})

	for _, dot := range p.Dots {
		pos := strings.IndexByte(s, dot.Byte)
		if pos < 0 {
			break
		}

		var val interface{} = s[:pos]
		s = s[pos+1:]

		if dot.Name == "-" || dot.Name == "" {
			continue
		}

		val, _ = dot.Converters.Convert(val)
		m[dot.Name] = val
	}

	return m
}

type Dot struct {
	Type
	Byte       byte
	Name       string
	Sample     string
	Converters Converters
}

func (d Dot) Valid() bool {
	return !(d.Name == "-" || d.Name == "")
}

var ErrBadPattern = errors.New("bad pattern")

var digitsRegexp = regexp.MustCompile(`^\d+$`)

type Type int

const (
	String Type = iota
	DateTime
	Float
	Digits
)

func NewPattern(sample, pattern string) (*Pattern, error) {
	if len(pattern) > len(sample) {
		return nil, ErrBadPattern
	}
	var dots []Dot

	for {
		pos := strings.IndexByte(pattern, '#')
		if pos < 0 {
			break
		}

		parts := split(strings.TrimSpace(pattern[:pos]), "|")
		name := parts[0]

		var converters []Converter
		typ := String

		dotSample := strings.TrimRight(sample[:pos], " ")
		if ss.ContainsAny(name, "time", "date") {
			converters = append(converters, TimeValue(dotSample))
			typ = DateTime
		} else if digitsRegexp.MatchString(dotSample) {
			converters = append(converters, DigitsValue())
			typ = Digits
		} else if strings.Count(dotSample, ".") == 1 &&
			digitsRegexp.MatchString(strings.ReplaceAll(dotSample, ".", "")) {
			converters = append(converters, FloatValue())
			typ = Float
		}

		for i := 1; i < len(parts); i++ {
			converters = append(converters, filters[parts[i]])
		}

		dot := Dot{Byte: sample[pos], Name: name, Converters: converters, Type: typ, Sample: dotSample}
		dots = append(dots, dot)
		pattern = pattern[pos+1:]
		sample = sample[pos+1:]
	}

	return &Pattern{Pattern: pattern, Dots: dots}, nil
}

func split(name, sep string) []string {
	var parts []string
	for _, v := range strings.Split(name, sep) {
		if v = strings.TrimSpace(v); v != "" {
			parts = append(parts, v)
		}
	}

	if len(parts) == 0 {
		return []string{name}
	}

	return parts
}

type Converter interface {
	Convert(v interface{}) (interface{}, error)
}

type Converters []Converter

func (c Converters) Convert(v interface{}) (interface{}, error) {
	for _, f := range c {
		if vv, err := f.Convert(v); err != nil {
			return v, err
		} else {
			v = vv
		}
	}

	return v, nil
}

func FloatValue() Converter             { return &floatValue{} }
func DigitsValue() Converter            { return &digitsValue{} }
func TimeValue(layout string) Converter { return &timeValue{layout: layout} }
func UriPath() Converter                { return &uriPath{} }

type uriPath struct{}
type timeValue struct{ layout string }
type digitsValue struct{}
type floatValue struct{}

func (t timeValue) Convert(v interface{}) (interface{}, error) {
	return time.Parse(t.layout, v.(string))
}

func (t digitsValue) Convert(v interface{}) (interface{}, error) { return strconv.Atoi(v.(string)) }

func (t floatValue) Convert(v interface{}) (interface{}, error) {
	return strconv.ParseFloat(v.(string), 64)
}
func (uriPath) Convert(v interface{}) (interface{}, error) {
	u, err := url.Parse(v.(string))
	if err != nil {
		return nil, err
	}

	return u.Path, nil
}
