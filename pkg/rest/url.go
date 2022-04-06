package rest

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/gg/pkg/osx"
	"github.com/bingoohuang/gg/pkg/rotate"
	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/bingoohuang/gg/pkg/timex"
)

var reScheme = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+-.]*://`)

type FixURIConfig struct {
	DefaultScheme string
	DefaultHost   string
	DefaultPort   int
	FatalErr      bool
}

func WithDefaultScheme(v string) FixURIConfigFn { return func(c *FixURIConfig) { c.DefaultScheme = v } }
func WithDefaultHost(v string) FixURIConfigFn   { return func(c *FixURIConfig) { c.DefaultHost = v } }
func WithDefaultPort(v int) FixURIConfigFn      { return func(c *FixURIConfig) { c.DefaultPort = v } }
func WithFatalErr(v bool) FixURIConfigFn        { return func(c *FixURIConfig) { c.FatalErr = v } }

type FixURIResult struct {
	Data string
	Err  error
}

func (r FixURIResult) OK() bool { return r.Err == nil }

func FixURI(uri string, fns ...FixURIConfigFn) (rr FixURIResult) {
	config := (FixURIConfigFns(fns)).Create()
	defer func() {
		if rr.Err != nil && config.FatalErr {
			log.Fatal(rr.Err)
		}
	}()

	if uri == ":" {
		uri = ":" + strconv.Itoa(config.DefaultPort)
	}

	// ex) :8080/hello or /hello or :
	if strings.HasPrefix(uri, ":") || strings.HasPrefix(uri, "/") {
		uri = config.DefaultHost + uri
	}

	// ex) example.com/hello
	if !reScheme.MatchString(uri) {
		uri = config.DefaultScheme + "://" + uri
	}

	u, err := url.Parse(uri)
	if err != nil {
		rr.Err = fmt.Errorf("parse %s failed: %s", uri, err)
		return rr
	}

	u.Host = strings.TrimSuffix(u.Host, ":")
	if u.Path == "" {
		u.Path = "/"
	}

	return FixURIResult{Data: u.String()}
}

func MaybeURL(out string) (string, bool) {
	if out == "stdout" {
		return "", false
	}

	if ss.HasPrefix(out, "http://", "https://") {
		return out, true
	}

	if ss.HasPrefix(out, ":", "/", ".") {
		uri := FixURI(out)
		return uri.Data, uri.OK()
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
		uri := FixURI(out)
		return uri.Data, uri.OK()
	}

	c := &rotate.Config{}
	if rotate.ParseOutputPath(c, out); c.Append || c.MaxSize > 0 {
		return "", false
	}

	uri := FixURI(out)
	return uri.Data, uri.OK()
}

type FixURIConfigFn func(*FixURIConfig)

type FixURIConfigFns []FixURIConfigFn

func (fns FixURIConfigFns) Create() *FixURIConfig {
	c := &FixURIConfig{
		DefaultScheme: "http",
		DefaultHost:   "127.0.0.1",
	}

	for _, f := range fns {
		f(c)
	}

	return c
}
