package rest

import (
	"github.com/bingoohuang/gg/pkg/osx"
	"github.com/bingoohuang/gg/pkg/rotate"
	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/bingoohuang/gg/pkg/timex"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var reScheme = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+-.]*://`)

const defaultScheme, defaultHost = "http", "127.0.0.1"

func FixURI(uri string) (string, error) {
	if uri == ":" {
		uri = ":80"
	}

	// ex) :8080/hello or /hello or :
	if strings.HasPrefix(uri, ":") || strings.HasPrefix(uri, "/") {
		uri = defaultHost + uri
	}

	// ex) example.com/hello
	if !reScheme.MatchString(uri) {
		uri = defaultScheme + "://" + uri
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	u.Host = strings.TrimSuffix(u.Host, ":")
	if u.Path == "" {
		u.Path = "/"
	}

	return u.String(), nil
}

func MaybeURL(out string) (string, bool) {
	if out == "stdout" {
		return "", false
	}

	if ss.HasPrefix(out, "http://", "https://") {
		return out, true
	}

	if ss.HasPrefix(out, ":", "/", ".") {
		uri, err := FixURI(out)
		return uri, err == nil
	}

	if osx.CanExpandHome(out) {
		return osx.ExpandHome(out), false
	}

	if ss.Contains(out, ".txt", ".log", ".gz", ".out", ".http", ".json") {
		return out, false
	}

	if fn := timex.FormatTime(time.Now(), out); fn != out {
		return "", false
	}

	// like ip:port
	if regexp.MustCompile(`^\d{1,3}((.\d){1,3}){3}(:\d+)?`).MatchString(out) {
		uri, err := FixURI(out)
		return uri, err == nil
	}

	c := &rotate.Config{}
	if rotate.ParseOutputPath(c, out); c.Append || c.MaxSize > 0 {
		return "", false
	}

	uri, err := FixURI(out)
	return uri, err == nil
}
