package ss

import (
	"strings"
)

// Repeat repeats s with separator seq for n times.
func Repeat(s, sep string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		if i == 0 {
			result = s
		} else {
			result += sep + s
		}
	}

	return result
}

// FirstWord returns the first word of the SQL statement s.
func FirstWord(s string) string {
	if v := FieldsN(strings.TrimSpace(s), 2); len(v) > 0 {
		return v[0]
	}

	return ""
}

func RemoveAll(s, old string) string { return strings.ReplaceAll(s, old, "") }

func Ori(a, b int) int {
	if a == 0 {
		return b
	}

	return a
}

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

func AnyOfFold(s string, ss ...string) bool {
	s = strings.ToLower(s)
	for _, of := range ss {
		if s == strings.ToLower(of) {
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

//func Jsonify(v interface{}) string {
//	b := &bytes.Buffer{}
//	encoder := json.NewEncoder(b)
//	encoder.SetEscapeHTML(false)
//	if err := encoder.Encode(v); err != nil {
//		return err.Error()
//	}
//
//	return b.String()
//}

func ToSet(v []string) map[string]bool {
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
	N           int
}

type SplitOption func(*SplitConfig)

func WithConfig(v SplitConfig) SplitOption { return func(c *SplitConfig) { *c = v } }
func WithSeps(v string) SplitOption        { return func(c *SplitConfig) { c.Separators = v } }
func WithN(v int) SplitOption              { return func(c *SplitConfig) { c.N = v } }
func WithIgnoreEmpty(v bool) SplitOption   { return func(c *SplitConfig) { c.IgnoreEmpty = v } }
func WithTrimSpace(v bool) SplitOption     { return func(c *SplitConfig) { c.TrimSpace = v } }
func WithCase(v Case) SplitOption          { return func(c *SplitConfig) { c.Case = v } }

func Split(s string, options ...SplitOption) []string {
	v := make([]string, 0)
	c := createConfig(options)

	ff := FieldsFuncN(s, c.N, func(r rune) bool {
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
	c := &SplitConfig{TrimSpace: true, IgnoreEmpty: true}
	for _, o := range options {
		o(c)
	}

	if c.Separators == "" {
		c.Separators = ", "
	}

	return c
}

func ToLowerKebab(name string) string {
	var sb strings.Builder

	isUpper := func(c uint8) bool { return 'A' <= c && c <= 'Z' }

	for i := 0; i < len(name); i++ {
		c := name[i]
		if isUpper(c) {
			if sb.Len() > 0 {
				if i+1 < len(name) && (!(i-1 >= 0 && isUpper(name[i-1])) || !isUpper(name[i+1])) {
					sb.WriteByte('-')
				}
			}
			sb.WriteByte(c - 'A' + 'a')
		} else {
			sb.WriteByte(c)
		}
	}

	return sb.String()
}
