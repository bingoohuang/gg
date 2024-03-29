package hlog

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Store defines the interface to Store a log.
type Store interface {
	// Store stores the log in database like MySQL, InfluxDB, and etc.
	Store(c *gin.Context, log *Log)
}

// Log describes info about HTTP request.
type Log struct {
	ID  string
	Biz string

	// Method is GET etc.
	Method string
	URL    string
	IPAddr string

	RspHeader http.Header
	ReqBody   string

	// RspStatus, like 200, 404.
	RspStatus int
	// ReqHeader records the response header.
	ReqHeader http.Header
	// RespSize is number of bytes of the response sent.
	RespSize int
	// RspBody is the response body(limit to 1000).
	RspBody string

	Created time.Time

	// Start records the start time of the request.
	Start time.Time
	// End records the end time of the request.
	End time.Time
	// Duration means how long did it take to.
	Duration time.Duration
	Attrs    Attrs

	Option     *Option
	PathParams gin.Params
	Request    *http.Request
}

func (l *Log) pathVar(name string) string {
	for _, p := range l.PathParams {
		if p.Key == name {
			return p.Value
		}
	}

	return ""
}

func (l *Log) pathVars() interface{} {
	m := make(map[string]string)

	for _, p := range l.PathParams {
		m[p.Key] = p.Value
	}

	return m
}

func (l *Log) queryVar(name string) string {
	return At(l.Request.URL.Query()[name], 0)
}

func (l *Log) queryVars() string {
	return l.Request.URL.Query().Encode()
}

func (l *Log) paramVar(name string) string {
	return At(l.Request.Form[name], 0)
}

func (l *Log) paramVars() string {
	return l.Request.Form.Encode()
}
