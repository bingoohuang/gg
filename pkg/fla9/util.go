package fla9

import (
	"io/ioutil"
	"log"
	"strings"
)

// ParseFileArg parse an argument which represents a string content,
// or @file to represents the file's content.
func ParseFileArg(arg string) (file string, data []byte) {
	if strings.HasPrefix(arg, "@") {
		f := (arg)[1:]
		if v, err := ioutil.ReadFile(f); err != nil {
			log.Fatalf("failed to read file %s, error: %v", f, err)
			return f, nil
		} else {
			return f, v
		}
	}

	return "", []byte(arg)
}
