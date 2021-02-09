package strcase

import (
	"strings"
)

// ToSnake converts a string to snake_case
func ToSnake(s string) string {
	return ToDelimited(s, '_')
}

// ToSnakeUpper converts a string to SCREAMING_SNAKE_CASE
func ToSnakeUpper(s string) string {
	return ToDelimitedScreaming(s, '_', true)
}

// ToKebab converts a string to kebab-case
func ToKebab(s string) string {
	return ToDelimited(s, '-')
}

// ToKebabUpper converts a string to SCREAMING-KEBAB-CASE
func ToKebabUpper(s string) string {
	return ToDelimitedScreaming(s, '-', true)
}

// ToDelimited converts a string to delimited.snake.case (in this case `del = '.'`)
func ToDelimited(s string, del uint8) string {
	return ToDelimitedScreaming(s, del, false)
}

// ToDelimitedUpper converts a string to SCREAMING.DELIMITED.SNAKE.CASE
// (in this case `del = '.'; screaming = true`) or delimited.snake.case (in this case `del = '.'; screaming = false`)
func ToDelimitedUpper(s string, del uint8) string {
	return ToDelimitedScreaming(s, del, true)
}

// ToDelimitedScreaming converts a string to SCREAMING.DELIMITED.SNAKE.CASE
// (in this case `del = '.'; screaming = true`) or delimited.snake.case (in this case `del = '.'; screaming = false`)
func ToDelimitedScreaming(s string, del uint8, screaming bool) string {
	s = addWordBoundariesToNumbers(s)
	s = strings.Trim(s, " ")
	n := ""

	for i, v := range s {
		// treat acronyms as words, eg for JSONData -> JSON is a whole word
		nextCaseIsChanged := false

		if i+1 < len(s) {
			next := s[i+1]
			if isCaseChanged(v, int32(next)) {
				nextCaseIsChanged = true
			}
		}

		switch {
		case i > 0 && n[len(n)-1] != del && nextCaseIsChanged:
			// add underscore if next letter case type is changed
			if IsA2Z(v) {
				n += string(del) + string(v)
			} else if Isa2z(v) {
				n += string(v) + string(del)
			}
		case anyOf(v, ' ', '_', '-'):
			if len(n) > 0 && n[len(n)-1] == del {
				continue
			}
			// replace spaces/underscores with delimiters
			n += string(del)
		default:
			n += string(v)
		}
	}

	if screaming {
		return strings.ToUpper(n)
	}

	return strings.ToLower(n)
}

func anyOf(v int32, oneOfs ...int32) bool {
	for _, one := range oneOfs {
		if v == one {
			return true
		}
	}

	return false
}

func isCaseChanged(v, next int32) bool {
	return IsA2Z(v) && Isa2z(next) ||
		Isa2z(v) && IsA2Z(next) ||
		Is029(v) && IsA2Z(next)
}

// Is029 tells v is 0-9
func Is029(v int32) bool {
	return v >= '0' && v <= '9'
}

// Isa2z tells v is a-z
func Isa2z(v int32) bool {
	return v >= 'a' && v <= 'z'
}

// IsA2Z tells v is A-Z
func IsA2Z(v int32) bool {
	return v >= 'A' && v <= 'Z'
}
