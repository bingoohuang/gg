package timex

import "regexp"

type javaFmtGoLayout struct {
	JavaRegex *regexp.Regexp
	GoLayout  string
}

var timeFormatConvert = []javaFmtGoLayout{
	{JavaRegex: regexp.MustCompile(`(?i)yyyy`), GoLayout: "2006"},
	{JavaRegex: regexp.MustCompile(`(?i)yy`), GoLayout: "06"},
	{JavaRegex: regexp.MustCompile(`MM`), GoLayout: "01"},
	{JavaRegex: regexp.MustCompile(`(?i)dd`), GoLayout: "02"},
	{JavaRegex: regexp.MustCompile(`(?i)hh`), GoLayout: "15"},
	{JavaRegex: regexp.MustCompile(`mm`), GoLayout: "04"},
	{JavaRegex: regexp.MustCompile(`(?i)sss`), GoLayout: "000"},
	{JavaRegex: regexp.MustCompile(`(?i)ss`), GoLayout: "05"},
}

// ConvertLayout converts Java style layout to golang.
func ConvertLayout(s string) string {
	for _, f := range timeFormatConvert {
		s = f.JavaRegex.ReplaceAllString(s, f.GoLayout)
	}

	return s
}
