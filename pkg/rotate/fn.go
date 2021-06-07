package rotate

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/bingoohuang/gg/pkg/timex"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// bufWriter is a Writer interface that also has a Flush method.
type bufWriter interface {
	io.Writer
	io.Closer
	Flush() error
}

type FileWriter struct {
	FnTemplate string
	MaxSize    uint64
	Append     bool

	file       *os.File
	curFn      string
	curSize    uint64
	rotateFunc func() bool
	writer     bufWriter
	DotGz      string
}

func NewFileWriter(fnTemplate string, maxSize uint64, append bool) *FileWriter {
	hasGz := strings.HasSuffix(fnTemplate, ".gz")
	dotGz := ""
	if hasGz {
		dotGz = ".gz"
		fnTemplate = strings.TrimSuffix(fnTemplate, ".gz")
	}
	r := &FileWriter{
		FnTemplate: fnTemplate,
		DotGz:      dotGz,
		MaxSize:    maxSize,
		Append:     append,
		rotateFunc: func() bool { return false },
	}

	if r.MaxSize > 0 {
		r.rotateFunc = func() bool { return r.curSize >= r.MaxSize }
	}

	return r
}

func (w *FileWriter) Write(p []byte) (int, error) {
	newFn := NewFilename(w.FnTemplate)

	for {
		fn, index := Filename(newFn, w.rotateFunc(), w.DotGz)
		if fn == w.curFn {
			break
		}

		if ok, err := w.openFile(fn, index); err != nil {
			return 0, err
		} else if ok {
			break
		}
	}

	n, err := w.writer.Write(p)
	w.curSize += uint64(n)
	return n, err
}

type gzipWriter struct {
	Buf *bufio.Writer
	*gzip.Writer
}

func (w *gzipWriter) Close() error {
	return w.Writer.Close()
}

func (w *gzipWriter) Flush() error {
	return w.Buf.Flush()
}

type bufioWriter struct {
	*bufio.Writer
}

func (b *bufioWriter) Close() error { return b.Writer.Flush() }
func (b *bufioWriter) Flush() error { return b.Writer.Flush() }

func (w *FileWriter) openFile(fn string, index int) (ok bool, err error) {
	_ = w.Close()
	if index == 2 { // rename bbb-2021-05-27-18-26.http to bbb-2021-05-27-18-26_00001.http
		_ = os.Rename(w.curFn+w.DotGz, SetFileIndex(w.curFn, 1)+w.DotGz)
	}

	w.file, err = os.OpenFile(fn+w.DotGz, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o660)
	if err != nil {
		return false, err
	}

	w.curFn = fn

	if w.DotGz != "" {
		gw := gzip.NewWriter(w.file)
		w.writer = &gzipWriter{
			Buf:    bufio.NewWriter(gw),
			Writer: gw,
		}
	} else {
		w.writer = &bufioWriter{
			Writer: bufio.NewWriter(w.file),
		}
	}

	ok = true

	if stat, _ := w.file.Stat(); stat != nil {
		if w.curSize = uint64(stat.Size()); w.curSize > 0 {
			ok = !w.rotateFunc()
		}
	}

	return ok, nil
}

type Flusher interface {
	Flush() error
}

func (w *FileWriter) Flush() error {
	if w.writer != nil {
		return w.writer.Flush()
	}

	return nil
}

func (w *FileWriter) Close() error {
	if w.writer != nil && w.file != nil {
		_ = w.writer.Close()
		_ = w.file.Close()
		w.writer = nil
		w.file = nil
	}
	return nil
}

func NewFilename(template string) string {
	fn := time.Now().Format(timex.ConvertLayout(template))
	fn = filepath.Clean(fn)
	_, fn = FindMaxFileIndex(fn, "")
	return fn
}

func Filename(fn string, rotate bool, dotGz string) (string, int) {
	if !rotate {
		return fn, 0
	}

	max, _ := FindMaxFileIndex(fn, dotGz)
	if max <= 0 {
		return fn, 0
	}

	n := max + 1
	return SetFileIndex(fn, n), n
}

func GetFileIndex(path string) int {
	_, index, _ := SplitBaseIndexExt(path)
	if index == "" {
		return -1
	}

	v, _ := strconv.Atoi(index)
	return v
}

func SetFileIndex(path string, index int) string {
	base, _, ext := SplitBaseIndexExt(path)
	return fmt.Sprintf("%s_%05d%s", base, index, ext)
}

// FindMaxFileIndex finds the max index of a file like log-2021-05-27_00001.log.
// return maxIndex = 0 there is no file matches log-2021-05-27*.log.
// return maxIndex >= 1 tell the max index in matches.
func FindMaxFileIndex(path string, dotGz string) (int, string) {
	base, _, ext := SplitBaseIndexExt(path)
	matches, _ := filepath.Glob(base + "*" + ext + dotGz)
	if len(matches) == 0 {
		return 0, path
	}

	maxIndex := 1
	maxFn := path
	for _, fn := range matches {
		fn = strings.TrimSuffix(fn, dotGz)
		if index := GetFileIndex(fn); index > maxIndex {
			maxIndex = index
			maxFn = fn
		}
	}

	return maxIndex, maxFn
}

var idx = regexp.MustCompile(`_\d{5,}`)

func SplitBaseIndexExt(path string) (base, index, ext string) {
	if subs := idx.FindAllStringSubmatchIndex(path, -1); len(subs) > 0 {
		sub := subs[len(subs)-1]
		return path[:sub[0]], path[sub[0]+1 : sub[1]], path[sub[1]:]
	}

	ext = filepath.Ext(path)
	return strings.TrimSuffix(path, ext), "", ext
}
