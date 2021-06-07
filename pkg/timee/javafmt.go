package timee

import "regexp"

var timeFormatConvert = map[*regexp.Regexp]string{
	regexp.MustCompile(`(?i)yyyy`): "2006",
	regexp.MustCompile(`(?i)yy`):   "06",
	regexp.MustCompile(`MM`):       "01",
	regexp.MustCompile(`(?i)dd`):   "02",
	regexp.MustCompile(`(?i)hh`):   "15",
	regexp.MustCompile(`mm`):       "04",
	regexp.MustCompile(`(?i)sss`):  "000",
	regexp.MustCompile(`(?i)ss`):   "05",
}

// ConvertLayout converts Java style layout to golang.
func ConvertLayout(s string) string {
	for r, f := range timeFormatConvert {
		s = r.ReplaceAllString(s, f)
	}

	return s
}
