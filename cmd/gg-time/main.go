package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		now := time.Now()
		printTime(now)
		return
	}

	for _, arg := range args {
		if regexp.MustCompile(`^\d+$`).MatchString(arg) {
			formats := []string{
				`20060102150405`,
				`200601021504`,
				`2006010215`,
				`20060102`,
			}
			found := false
			for _, f := range formats {
				t, err := time.ParseInLocation(f, arg, time.Local)
				if err == nil {
					printTime(t)
					found = true
					break
				}
			}

			if !found {
				v, _ := strconv.ParseInt(arg, 10, 64)
				fmt.Println("as unix:\t", time.Unix(v, 0).Format(time.RFC3339))
				fmt.Println("as unix milli:\t", time.UnixMilli(v).Format(time.RFC3339))
			}
		} else {
			t, err := time.Parse(time.RFC3339, arg)
			if err == nil {
				printTime(t)
				continue
			}
			formats := []string{
				`2006-01-02T15:04:05`,
				`2006-01-02T15:04`,
				`2006-01-02T15`,
				`2006-01-02 15:04:05`,
				`2006-01-02 15:04`,
				`2006-01-02 15`,
				`2006-01-02`,
			}
			for _, f := range formats {
				t, err = time.ParseInLocation(f, arg, time.Local)
				if err == nil {
					printTime(t)
					break
				}
			}
		}
	}
}

func printTime(now time.Time) {
	fmt.Println("now:\t\t", now.Format(time.RFC3339))
	fmt.Println("unix:\t\t", now.Unix())
	fmt.Println("unix milli:\t", now.UnixMilli())
	fmt.Println("unix micro:\t", now.UnixMicro())
	fmt.Println("unix nano:\t", now.UnixNano())
}
