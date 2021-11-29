package osx

import (
	"fmt"
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
