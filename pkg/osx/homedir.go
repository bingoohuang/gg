package osx

import (
	"os"
	"path/filepath"
	"strings"
)

var Home string

func init() {
	Home, _ = os.UserHomeDir()
}

func ExpandHome(s string) bool {
	return strings.HasPrefix(s, "~")
}
func Expand(s string) string {
	if ExpandHome(s) {
		return filepath.Join(Home, s[1:])
	}

	return s
}
