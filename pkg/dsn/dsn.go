package dsn

import (
	"fmt"
	"strconv"
	"strings"
)

// Flag is the datasource dsn structure.
type Flag struct {
	Username, Password, Host string
	Port                     int
	Database                 string
}

// ParseFlag parses format like user:pass@host:port/db.
func ParseFlag(source string) (dc *Flag, err error) {
	dc = &Flag{}
	atPos := strings.LastIndex(source, "@")
	if atPos < 0 {
		err = fmt.Errorf("invalid source: %s, should be username:password@host:port", source)
		return
	}

	userPart := source[:atPos]
	if n := strings.IndexAny(userPart, "/:"); n > 0 {
		if dc.Username = userPart[:n]; n+1 < len(userPart) {
			dc.Password = userPart[n+1:]
		}
	} else {
		dc.Username = userPart
	}

	if atPos+1 >= len(source) {
		err = fmt.Errorf("invalid source: %s, should be username:password@host:port", source)
		return
	}

	hostPart := source[atPos+1:]

	if n := strings.LastIndex(hostPart, "/"); n > 0 {
		dc.Database = hostPart[n+1:]
		hostPart = hostPart[:n]
	}

	if n := strings.LastIndex(hostPart, ":"); n > 0 {
		if dc.Host = hostPart[:n]; n+1 < len(hostPart) {
			dc.Port, err = strconv.Atoi(hostPart[n+1:])
			if err != nil {
				err = fmt.Errorf("port %s is not a number", hostPart[n+1:])
				return
			}
		}
	} else {
		dc.Host = hostPart
	}

	return
}
