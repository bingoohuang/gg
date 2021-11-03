package ss

import "strconv"

// FormatFloat format a float64 to a string with the fewest digits necessary.
func FormatFloat(f float64) string {
	// The -1 as the 3rd parameter tells the func to print the fewest digits necessary to accurately represent the float.
	// See here: https://golang.org/pkg/strconv/#FormatFloat
	return strconv.FormatFloat(f, 'f', -1, 64)
}
