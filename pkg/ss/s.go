package ss

import (
	"bytes"
	"encoding/json"
	"strings"
)

func RemoveAll(s, old string) string { return strings.ReplaceAll(s, old, "") }

func Or(a, b string) string {
	if a == "" {
		return b
	}

	return a
}

func Ifi(b bool, s1, s2 int) int {
	if b {
		return s1
	}

	return s2
}

func If(b bool, s1, s2 string) string {
	if b {
		return s1
	}

	return s2
}

func ContainsFold(s string, ss ...string) bool {
	s = strings.ToLower(s)
	for _, of := range ss {
		of = strings.ToLower(of)
		if strings.Contains(s, of) {
			return true
		}
	}
	return false
}

func Contains(s string, ss ...string) bool {
	for _, of := range ss {
		if strings.Contains(s, of) {
			return true
		}
	}
	return false
}

func AnyOf(s string, ss ...string) bool {
	for _, of := range ss {
		if s == of {
			return true
		}
	}
	return false
}

func HasPrefix(s string, ss ...string) bool {
	for _, one := range ss {
		if strings.HasPrefix(s, one) {
			return true
		}
	}
	return false
}

func HasSuffix(s string, ss ...string) bool {
	for _, one := range ss {
		if strings.HasSuffix(s, one) {
			return true
		}
	}
	return false
}

func Jsonify(v interface{}) string {
	b := &bytes.Buffer{}
	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return err.Error()
	}

	return b.String()
}

func ToMap(v []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range v {
		m[s] = true
	}
	return m
}

type Case int

const (
	CaseUnchanged Case = iota
	CaseLower
	CaseUpper
)

type SplitConfig struct {
	Separators  string
	IgnoreEmpty bool
	TrimSpace   bool
	Case        Case
}

type SplitOption func(*SplitConfig)

func WithConfig(v SplitConfig) SplitOption { return func(c *SplitConfig) { *c = v } }
func WithSeparators(v string) SplitOption  { return func(c *SplitConfig) { c.Separators = v } }
func WithIgnoreEmpty() SplitOption         { return func(c *SplitConfig) { c.IgnoreEmpty = true } }
func WithTrimSpace() SplitOption           { return func(c *SplitConfig) { c.TrimSpace = true } }
func WithUpper() SplitOption               { return func(c *SplitConfig) { c.Case = CaseUpper } }
func WithLower() SplitOption               { return func(c *SplitConfig) { c.Case = CaseLower } }

func Split(s string, options ...SplitOption) []string {
	v := make([]string, 0)
	c := createConfig(options)

	ff := strings.FieldsFunc(s, func(r rune) bool {
		return strings.ContainsRune(c.Separators, r)
	})

	for _, f := range ff {
		if c.TrimSpace {
			f = strings.TrimSpace(f)
		}
		if c.IgnoreEmpty && f == "" {
			continue
		}
		switch c.Case {
		case CaseLower:
			f = strings.ToLower(f)
		case CaseUpper:
			f = strings.ToUpper(f)
		}

		v = append(v, f)
	}

	return v
}

func createConfig(options []SplitOption) *SplitConfig {
	c := &SplitConfig{}
	for _, o := range options {
		o(c)
	}

	if c.Separators == "" {
		c.Separators = ", "
	}

	return c
}
