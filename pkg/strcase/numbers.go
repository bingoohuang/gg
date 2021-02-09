package strcase

import (
	"regexp"
)

// nolint gochecknoglobals
var (
	numberSequence    = regexp.MustCompile(`([a-zA-Z]\d+)([a-zA-Z]?)`)
	numberReplacement = []byte(`$1 $2`)
)

func addWordBoundariesToNumbers(s string) string {
	return string(numberSequence.ReplaceAll([]byte(s), numberReplacement))
}
