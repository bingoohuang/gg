package rest

import (
	"github.com/bingoohuang/gg/pkg/ss"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

func LogResponse(r *http.Response, verbose string) {
	if r != nil && ss.Contains(verbose, "rsp", "all") {
		if dump, err := httputil.DumpResponse(r, true); err != nil {
			log.Printf("E! Failed to dump response: %v", err)
		} else {
			log.Printf("Dumped response: %s", dump)
		}
	}
}

func LogRequest(r *http.Request, verbose string) {
	if r != nil && ss.Contains(verbose, "req", "all") {
		if dump, err := httputil.DumpRequest(r, true); err != nil {
			log.Printf("Failed to dump request: %v", err)
		} else {
			log.Printf("Dumped request: %s", dump)
		}
	}
}

func ReadCloseBody(r *http.Response) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	if r.Body == nil {
		return nil, nil
	}

	data, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	if err != nil {
		return nil, err
	}

	return data, nil
}

func DiscardCloseBody(r *http.Response) error {
	if r == nil {
		return nil
	}
	if r.Body == nil {
		return nil
	}
	_, err := io.Copy(io.Discard, r.Body)
	r.Body.Close()
	return err
}
