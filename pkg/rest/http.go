package rest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

type Rest struct {
	Body             []byte
	Method, Addr     string
	Headers          map[string]string
	Result           interface{}
	DisableKeepAlive bool
	Context          context.Context
	Timeout          time.Duration
}

// Post execute HTTP POST request.
func (r *Rest) Post() (*Rsp, error) { return r.do("POST") }

// Get execute HTTP GET request.
func (r *Rest) Get() (*Rsp, error) { return r.do("GET") }

// Delete execute HTTP GET request.
func (r *Rest) Delete() (*Rsp, error) { return r.do("DELETE") }

// Upload execute HTTP GET request.
func (r *Rest) Upload(filename string, fileData []byte) (*Rsp, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	_, _ = part.Write(fileData)
	_ = writer.Close()

	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}

	r.Body = body.Bytes()
	r.Headers["Content-Type"] = writer.FormDataContentType()
	r.Method = "POST"
	return r.Do()
}

type Rsp struct {
	Body   []byte
	Status int
	Header http.Header
	Cost   time.Duration
}

var Client = &http.Client{
	//Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		//DisableKeepAlives: true,
	},
}

// Do execute HTTP method request.
func (r *Rest) Do() (*Rsp, error) {
	return r.do(If(r.Method == "", "GET", r.Method))
}

// Do execute HTTP method request.
func (r *Rest) do(method string) (*Rsp, error) {
	var ctx context.Context
	if r.Context != nil {
		ctx = r.Context
	} else {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), DurationOr(r.Timeout, 10*time.Second))
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, method, r.Addr, bytes.NewReader(r.Body))
	if err != nil {
		return nil, err
	}
	req.Close = r.DisableKeepAlive
	if r != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", If(IsJSONBytes(r.Body),
			"application/json; charset=utf-8", "text/plain; charset=utf-8"))
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := Client.Do(req)
	cost := time.Since(start)
	if err != nil {
		return nil, err
	}

	rsp := &Rsp{Status: resp.StatusCode, Header: resp.Header, Cost: cost}
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rsp, err
	}
	_ = resp.Body.Close()

	rsp.Body = bodyData
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if r.Result != nil {
			_ = json.Unmarshal(bodyData, r.Result)
		}
		return rsp, nil
	}

	return rsp, errors.Wrapf(err, "status:%d", resp.StatusCode)
}

func DurationOr(a, b time.Duration) time.Duration {
	if a == 0 {
		return b
	}

	return a
}

// If tests condition to return a or b.
func If(condition bool, a, b string) string {
	if condition {
		return a
	}

	return b
}

// IsJSONBytes tests bytes b is in JSON format.
func IsJSONBytes(b []byte) bool {
	if len(b) == 0 {
		return false
	}

	var m interface{}
	return json.Unmarshal(b, &m) == nil
}
