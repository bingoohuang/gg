package logx

import "log"

func Fatalf(err error, format string, args ...interface{}) {
	if err != nil {
		log.Fatalf(format, args...)
	}
}
