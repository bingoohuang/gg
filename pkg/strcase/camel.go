package strcase

import (
	"strings"
)

// Converts a string to CamelCase
func toCamelInitCase(s string, initCase bool) string {
	s = addWordBoundariesToNumbers(s)
	s = strings.Trim(s, " ")
	n := ""
	capNext := initCase
	lastUpper := false

	for _, v := range s {
		if IsA2Z(v) {
			if lastUpper {
				n += strings.ToLower(string(v))
			} else {
				n += string(v)
				lastUpper = true
			}
		} else {
			lastUpper = false
		}

		if Is029(v) {
			n += string(v)
		}

		if Isa2z(v) {
			if capNext {
				n += strings.ToUpper(string(v))
			} else {
				n += string(v)
			}
		}

		capNext = anyOf(v, '_', ' ', '-')
	}

	return n
}

// ToCamel converts a string to CamelCase
func ToCamel(s string) string {
	return toCamelInitCase(s, true)
}

// ToCamelLower converts a string to lowerCamelCase
func ToCamelLower(s string) string {
	if s == "" {
		return s
	}

	i := 0
	for ; i < len(s); i++ {
		if r := rune(s[i]); !(r >= 'A' && r <= 'Z') {
			break
		}
	}

	if i == len(s) {
		return strings.ToLower(s)
	}

	if i > 1 { // nolint gomnd
		s = strings.ToLower(s[:i-1]) + s[i-1:]
	} else if i > 0 {
		s = strings.ToLower(s[:1]) + s[1:]
	}

	return toCamelInitCase(s, false)
}
