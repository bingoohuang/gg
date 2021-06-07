package timee

import (
	"errors"
	"strconv"
	"strings"
	"time"

	perrors "github.com/pkg/errors"
)

// ErrUnknownTimeFormat defines the error type for unknown time format.
var ErrUnknownTimeFormat = errors.New("unknown errors time format")

// JSONTime defines a time.Time that can be used in struct tag for JSON unmarshalling.
type JSONTime time.Time

// UnmarshalJSON unmarshals bytes to JSONTime.
func (t *JSONTime) UnmarshalJSON(b []byte) error {
	v, _ := TryUnQuoted(string(b))
	if v == "" {
		return nil
	}

	// 首先看是否是数字，表示毫秒数或者纳秒数
	if p, err := strconv.ParseInt(v, 10, 64); err == nil {
		*t = JSONTime(time.Unix(0, p*1000000)) // milliseconds range, 1 millis = 1000,000 nanos)
		return nil
	}

	v = strings.ReplaceAll(v, ",", ".")
	v = strings.ReplaceAll(v, "T", " ")
	v = strings.ReplaceAll(v, "-", "")
	v = strings.ReplaceAll(v, ":", "")
	v = strings.TrimSuffix(v, "Z")

	for _, f := range []string{"20060102 150405.000000", "20060102 150405.000"} {
		if tt, err := time.ParseInLocation(f, v, time.Local); err == nil {
			*t = JSONTime(tt)
			return nil
		}
	}

	return perrors.Wrapf(ErrUnknownTimeFormat, "value %s has unknown time format"+v)
}

// TryUnQuoted tries to unquote string.
func TryUnQuoted(v string) (string, bool) {
	vlen := len(v)
	yes := vlen >= 2 && v[0] == '"' && v[vlen-1] == '"'
	if !yes {
		return v, false
	}

	return v[1 : vlen-1], true
}
