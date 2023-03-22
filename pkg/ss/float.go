package ss

import "strconv"

// FormatFloat format a float64 to a string with the fewest digits necessary.
func FormatFloat(f float64) string {
	// The -1 as the 3rd parameter tells the func to print the fewest digits necessary to accurately represent the float.
	// See here: https://golang.org/pkg/strconv/#FormatFloat
	return strconv.FormatFloat(f, 'f', -1, 64)
}

type FormatFloatOptionConfig struct {
	// Prec controls the number of digits (excluding the exponent)
	// printed by the 'e', 'E', 'f', 'g', 'G', 'x', and 'X' formats.
	// For 'e', 'E', 'f', 'x', and 'X', it is the number of digits after the decimal point.
	// For 'g' and 'G' it is the maximum number of significant digits (trailing zeros are removed).
	// The special precision -1 uses the smallest number of digits
	Prec                int
	RemoveTrailingZeros bool
}

type FormatFloatOptionConfigFn func(*FormatFloatOptionConfig)

func WithRemoveTrailingZeros(yes bool) FormatFloatOptionConfigFn {
	return func(c *FormatFloatOptionConfig) {
		c.RemoveTrailingZeros = yes
	}
}

func WithPrec(prec int) FormatFloatOptionConfigFn {
	return func(c *FormatFloatOptionConfig) {
		c.Prec = prec
	}
}

func FormatFloatOption(f float64, fns ...FormatFloatOptionConfigFn) string {
	c := &FormatFloatOptionConfig{Prec: -1}
	for _, fn := range fns {
		fn(c)
	}

	b := strconv.FormatFloat(f, 'f', c.Prec, 64)
	if c.RemoveTrailingZeros && c.Prec != 0 {
		i := len(b) - 1
		for ; i >= 0; i-- {
			if b[i] != '0' {
				if b[i] != '.' {
					i++
				}
				break
			}
		}

		b = b[:i]
	}

	return b
}
