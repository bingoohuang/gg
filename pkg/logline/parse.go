package logline

import (
	"errors"
	"github.com/bingoohuang/gg/pkg/ss"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Converter interface {
	Convert(v interface{}) (interface{}, error)
}

var filters = map[string]Converter{
	"path": UriPath(),
}

type Pattern struct {
	Pattern    string
	Converters map[string]Converter
	Dots       []Dot
}

func (p Pattern) Parse(s string) map[string]interface{} {
	m := make(map[string]interface{})
	var err error

	for _, dot := range p.Dots {
		pos := strings.IndexByte(s, dot.Byte)
		if pos < 0 {
			break
		}

		ss := s[:pos]
		var val interface{} = ss
		s = s[pos+1:]
		if dot.Name == "-" || dot.Name == "" {
			continue
		}

		if len(dot.Converters) > 0 {
			for _, c := range dot.Converters {
				val, err = c.Convert(val)
				if err != nil {
					break
				}
			}
		} else if dot.Digits {
			if v, err := strconv.Atoi(ss); err == nil {
				m[dot.Name] = v
				continue
			}
		}
		m[dot.Name] = val
	}

	return m
}

type Dot struct {
	Byte       byte
	Digits     bool
	Name       string
	Sample     string
	Converters []Converter
}

var ErrBadPattern = errors.New("bad pattern")

var digitsRegexp = regexp.MustCompile(`^\d+$`)

func NewPattern(sample, s string) (*Pattern, error) {
	if len(s) > len(sample) {
		return nil, ErrBadPattern
	}
	var dots []Dot

	for {
		pos := strings.IndexByte(s, '#')
		if pos < 0 {
			break
		}

		var converters []Converter

		name := strings.TrimSpace(s[:pos])
		if name != "" {
			var parts []string
			for _, v := range strings.Split(name, "|") {
				v = strings.TrimSpace(v)
				if v != "" {
					parts = append(parts, v)
				}
			}
			if len(parts) > 0 {
				name = parts[0]

				for i := 1; i < len(parts); i++ {
					converters = append(converters, filters[parts[i]])
				}
			}
		}

		dotSample := strings.TrimRight(sample[:pos], " ")
		if ss.ContainsAny(name, "time", "date") && len(converters) == 0 {
			converters = []Converter{TimeValue(dotSample)}
		}

		digits := digitsRegexp.MatchString(dotSample)
		dot := Dot{
			Byte:       sample[pos],
			Name:       name,
			Sample:     dotSample,
			Digits:     digits,
			Converters: converters,
		}
		dots = append(dots, dot)

		s = s[pos+1:]
		sample = sample[pos+1:]
	}

	return &Pattern{Pattern: s, Dots: dots}, nil
}

type uriPath struct{}

func (uriPath) Convert(v interface{}) (interface{}, error) {
	u, err := url.Parse(v.(string))
	if err != nil {
		return nil, err
	}

	return u.Path, nil
}

func UriPath() Converter { return &uriPath{} }

type timeValue struct{ layout string }

func (t timeValue) Convert(v interface{}) (interface{}, error) {
	return time.Parse(t.layout, v.(string))
}

func TimeValue(layout string) Converter { return &timeValue{layout: layout} }
