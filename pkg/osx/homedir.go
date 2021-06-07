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

func CanExpandHome(s string) bool {
	return strings.HasPrefix(s, "~")
}
func ExpandHome(s string) string {
	if CanExpandHome(s) {
		return filepath.Join(Home, s[1:])
	}

	return s
}
