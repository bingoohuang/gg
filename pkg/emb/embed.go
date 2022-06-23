package emb

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// from https://github.com/pyros2097/go-embed

func AssetString(f fs.FS, name string, useGzip bool) string {
	return string(AssetBytes(f, name, useGzip))
}

func AssetBytes(f fs.FS, name string, useGzip bool) (data []byte) {
	data, _, _, _ = Asset(f, name, useGzip)
	return data
}

// Asset Gets the file from embed.FS if debug otherwise gets it from the stored
// data returns the data, the md5 hash of its content and its content type
// and error if it is not found.
func Asset(f fs.FS, name string, useGzip bool) (data []byte, hash, contentType string, err error) {
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}

	var fn fs.File
	fn, err = f.Open(name)
	if err != nil {
		return nil, "", "", err

	}
	data, err = io.ReadAll(fn)
	if err != nil {
		return nil, "", "", err
	}

	if useGzip && len(data) > 0 {
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		_, _ = w.Write(data)
		_ = w.Close()
		data = b.Bytes()
	}

	sum := md5.Sum(data)
	contentType = mime.TypeByExtension(filepath.Ext(name))
	return data, hex.EncodeToString(sum[:]), contentType, nil
}

func FileHandler(f fs.FS, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { ServeFile(f, name, w, r) }
}

func ServeFile(f fs.FS, name string, w http.ResponseWriter, r *http.Request) {
	ServeFileGzip(f, name, w, r, true)
}

func ServeFileGzip(f fs.FS, name string, w http.ResponseWriter, r *http.Request, useGzip bool) {
	if useGzip && w.Header().Get("Content-Encoding") == "gzip" {
		useGzip = false // if already gzip outside
	}

	data, hash, contentType, err := Asset(f, name, useGzip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("If-None-Match") == hash {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if useGzip {
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Add("Cache-Control", "public, max-age=31536000")
	w.Header().Add("ETag", hash)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
