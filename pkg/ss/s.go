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

func ContainsAnyFold(s string, ss ...string) bool {
	s = strings.ToLower(s)
	for _, of := range ss {
		of = strings.ToLower(of)
		if strings.Contains(s, of) {
			return true
		}
	}
	return false
}

func ContainsAny(s string, ss ...string) bool {
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
