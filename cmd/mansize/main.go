package main

import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/man"
	"os"
	"regexp"
	"strconv"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("usage: mansize 123445 12M")
	}

	for _, arg := range args {
		if regexp.MustCompile(`^\d+$`).MatchString(arg) {
			v, _ := strconv.ParseUint(os.Args[1], 10, 64)
			fmt.Println(arg, "=>", man.IBytes(v), "/", man.Bytes(v))
		} else {
			bytes, err := man.ParseBytes(arg)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(arg, "=>", bytes)
			}
		}
	}
}
