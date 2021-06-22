package timex

import (
	"regexp"
	"strings"
	"time"
)

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

// FormatTime format time with Java style layout.
func FormatTime(t time.Time, s string) string {
	for _, f := range timeFormatConvert {
		s = f.JavaRegex.ReplaceAllStringFunc(s, func(layout string) string {
			return t.Format(f.GoLayout)
		})
	}

	return s
}

// GlobName format time with Java style layout.
func GlobName(s string) string {
	for _, f := range timeFormatConvert {
		s = f.JavaRegex.ReplaceAllStringFunc(s, func(layout string) string {
			return strings.Repeat("?", len(layout))
		})
	}

	return s
}

// ConvertFormat converts Java style layout to golang.
func ConvertFormat(s string) string {
	for _, f := range timeFormatConvert {
		s = f.JavaRegex.ReplaceAllString(s, f.GoLayout)
	}

	return s
}
