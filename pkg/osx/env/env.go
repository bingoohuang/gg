package env

import (
	"github.com/bingoohuang/gg/pkg/man"
	"log"
	"os"
	"strconv"
	"strings"
)

func Bool(envName string, defaultValue bool) (value bool) {
	switch s := os.Getenv(envName); strings.ToLower(s) {
	case "yes", "y", "1", "on", "true", "t":
		return true
	}
	return defaultValue
}

func Int(envName string, defaultValue int) int {
	if s := os.Getenv(envName); s != "" {
		if size, err := strconv.Atoi(s); err != nil {
			log.Printf("parse env %s=%s failed: %+v", envName, s, err)
		} else {
			return size
		}
	}
	return defaultValue
}

func Size(envName string, defaultValue int) int {
	if s := os.Getenv(envName); s != "" {
		if size, err := man.ParseBytes(s); err != nil {
			log.Printf("parse env %s=%s failed: %+v", envName, s, err)
		} else if size >= 0 {
			return int(size)
		}
	}
	return defaultValue
}
