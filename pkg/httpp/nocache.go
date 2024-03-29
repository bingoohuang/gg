package httpp

import (
	"net/http"
	"time"
)

var epoch = time.Unix(0, 0).Format(time.RFC1123)

var noCacheHeaders = map[string]string{
	"Expires":         epoch,
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

var etagHeaders = []string{
	"ETag",
	"If-Modified-Since",
	"If-Match",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
}

// NoCache set no cache headers for http response.
// copy from https://stackoverflow.com/questions/33880343/go-webserver-dont-cache-files-using-timestamp
func NoCache(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		NoCacheHeaders(w, r)

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func NoCacheHeaders(w http.ResponseWriter, r *http.Request) {
	// Delete any ETag headers that may have been set
	for _, v := range etagHeaders {
		if r.Header.Get(v) != "" {
			r.Header.Del(v)
		}
	}

	// Set our NoCache headers
	for k, v := range noCacheHeaders {
		w.Header().Set(k, v)
	}
}
