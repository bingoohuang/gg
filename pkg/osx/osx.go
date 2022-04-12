package osx

import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/man"
	"log"
	"os"
)

func ExitIfErr(err error) {
	if err != nil {
		Exit(err.Error(), 1)
	}
}

func Exit(msg string, code int) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}

func EnvSize(envName string, defaultValue int) int {
	if s := os.Getenv(envName); s != "" {
		if size, err := man.ParseBytes(s); err != nil {
			log.Printf("parse env %s=%s failed: %+v", envName, s, err)
		} else if size > 0 {
			return int(size)
		}
	}
	return defaultValue
}
