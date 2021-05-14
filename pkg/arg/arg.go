package arg

import (
	"log"
	"os"
	"strings"
)

// ArgData returns argument s if it starts with @filename, the file contents will be replaced as the data.
func ArgData(s string) string {
	if !strings.HasPrefix(s, "@") {
		return s
	}

	data, err := os.ReadFile(s[1:])
	if err != nil {
		log.Fatalf("failed to read %s: %v", s, err)
	}
	return string(data)
}
