package logline

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/bingoohuang/gg/pkg/ss"
)

var filters = map[string]Converter{
	"path":     UriPath(),
	"duration": DurationPath(),
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

func (p Pattern) ParseBytes(s []byte) (map[string]interface{}, bool) {
	return p.Parse(SliceToString(s))
}

func (p Pattern) Parse(s string) (m map[string]interface{}, ok bool) {
	m = make(map[string]interface{})
	count := 0
	for _, dot := range p.Dots {
		if dot.EOF {
			if dot.Valid() {
				val, _ := dot.Converters.Convert(s)
				m[dot.Name] = val
			}
			count++
			break
		}
		pos := strings.Index(s, dot.Anchor)
		if pos < 0 {
			break
		}

		count++
		if dot.Valid() {
			val, _ := dot.Converters.Convert(s[:pos])
			m[dot.Name] = val
		}

		s = s[pos+len(dot.Anchor):]
	}

	ok = count == len(p.Dots)

	return
}

type Dot struct {
	Type
	Anchor     string
	Name       string
	Sample     string
	Converters Converters
	EOF        bool
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

func WithReplace(pairs ...string) func(*Option) {
	return func(option *Option) {
		option.Replaces = pairs
	}
}

type Option struct {
	Replaces []string
}

func (o Option) Replace(s string) string {
	for i := 0; i+1 < len(o.Replaces); i += 2 {
		s = strings.ReplaceAll(s, o.Replaces[i], o.Replaces[i+1])
	}

	return s
}

type (
	OptionFn  func(*Option)
	OptionFns []OptionFn
)

func (fs OptionFns) Apply(option *Option) {
	for _, f := range fs {
		f(option)
	}
}

func NewPattern(sample, pattern string, options ...OptionFn) (*Pattern, error) {
	var dots []Dot

	option := &Option{}
	OptionFns(options).Apply(option)
	sample = option.Replace(sample)
	for {
		pos := strings.IndexByte(pattern, '#')
		if pos < 0 && pattern == "" {
			break
		}

		left, leftSample := "", ""
		var anchor string

		if pos < 0 {
			left = pattern
			leftSample = sample
		} else {
			more := nextCrosses(pos, pattern)

			if pos+more >= len(sample) {
				return nil, ErrBadPattern
			}

			anchor = sample[pos : pos+more+1]
			left = pattern[:pos]
			leftSample = sample[:pos]
			pos += more
		}
		parts := split(strings.TrimSpace(left), "|")
		name := parts[0]

		var converters Converters
		typ := String

		dotSample := strings.Trim(leftSample, " ")
		if ss.Contains(name, "time", "date") {
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

		if typ == String && len(converters) > 0 {
			if v, err := converters.Convert(dotSample); err == nil {
				switch v.(type) {
				case float64, float32:
					typ = Float
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					typ = Digits
				}
			}
		}

		dot := Dot{Anchor: anchor, Name: name, Converters: converters, Type: typ, Sample: dotSample, EOF: pos < 0}
		dots = append(dots, dot)

		if pos < 0 {
			break
		}
		pattern = pattern[pos+1:]
		sample = sample[pos+1:]
	}

	return &Pattern{Pattern: pattern, Dots: dots}, nil
}

func nextCrosses(pos int, pattern string) int {
	for i := pos + 1; i < len(pattern); i++ {
		if pattern[i] != '#' {
			return i - pos - 1
		}
	}
	return 0
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
func DurationPath() Converter           { return &durationPath{} }

type (
	durationPath struct{}
	uriPath      struct{}
	timeValue    struct{ layout string }
	digitsValue  struct{}
	floatValue   struct{}
)

func (t timeValue) Convert(v interface{}) (interface{}, error) {
	return time.Parse(t.layout, v.(string))
}

func (t digitsValue) Convert(v interface{}) (interface{}, error) {
	vs := v.(string)
	if vs == "" || vs == "-" {
		return 0, nil
	}
	return strconv.Atoi(vs)
}

func (t floatValue) Convert(v interface{}) (interface{}, error) {
	vs := v.(string)
	if vs == "" || vs == "-" {
		return float64(0), nil
	}
	return strconv.ParseFloat(vs, 64)
}

func (uriPath) Convert(v interface{}) (interface{}, error) {
	u, err := url.Parse(v.(string))
	if err != nil {
		return nil, err
	}

	return u.Path, nil
}

func (d durationPath) Convert(v interface{}) (interface{}, error) {
	du, err := time.ParseDuration(v.(string))
	if err != nil {
		return v, err
	}

	return du.Seconds(), nil
}
