package iox

import (
	"io/ioutil"
	"os"
)

// WriteTempFile writes the content to a temporary file.
func WriteTempFile(fns ...WriteTempFileOptionFn) *WriteTempFileResult {
	o := &WriteTempFileOption{}
	for _, fn := range fns {
		fn(o)
	}
	if o.TempDir == "" {
		o.TempDir = os.TempDir()
	}

	r := o.writeTempFile()
	if r.Err != nil && o.PanicOnError {
		panic(r.Err)
	}

	return r
}

type WriteTempFileResult struct {
	Name string
	Err  error
}

type WriteTempFileOption struct {
	Content      []byte
	TempDir      string
	Pattern      string
	PanicOnError bool
}

func (o WriteTempFileOption) writeTempFile() *WriteTempFileResult {
	r := &WriteTempFileResult{}
	f, err := ioutil.TempFile(o.TempDir, o.Pattern)
	if err != nil {
		r.Err = err
		return r
	}

	if _, err := f.Write(o.Content); err != nil {
		r.Err = err
		return r
	}

	if err := f.Close(); err != nil {
		r.Err = err
		return r
	}

	r.Name = f.Name()
	return r
}

type WriteTempFileOptionFn func(*WriteTempFileOption)

func PanicOnError(c bool) WriteTempFileOptionFn {
	return func(o *WriteTempFileOption) {
		o.PanicOnError = c
	}
}

func WithTempPattern(c string) WriteTempFileOptionFn {
	return func(o *WriteTempFileOption) {
		o.Pattern = c
	}
}
func WithTempDir(c string) WriteTempFileOptionFn {
	return func(o *WriteTempFileOption) {
		o.TempDir = c
	}
}
func WithTempString(c string) WriteTempFileOptionFn {
	return func(o *WriteTempFileOption) {
		o.Content = []byte(c)
	}
}

func WithTempContent(c []byte) WriteTempFileOptionFn {
	return func(o *WriteTempFileOption) {
		o.Content = c
	}
}
