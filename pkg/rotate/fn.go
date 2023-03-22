package rotate

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/gg/pkg/timex"
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

	file        *os.File
	curFn       string
	curSize     uint64
	rotateFunc  func() bool
	writer      bufWriter
	DotGz       string
	maxIndex    int
	timedFn     string
	MaxKeepDays int
}

func NewFileWriter(fnTemplate string, maxSize uint64, append bool, maxKeepDays int) *FileWriter {
	hasGz := strings.HasSuffix(fnTemplate, ".gz")
	dotGz := ""
	if hasGz {
		dotGz = ".gz"
		fnTemplate = strings.TrimSuffix(fnTemplate, ".gz")
	}
	r := &FileWriter{
		FnTemplate:  fnTemplate,
		DotGz:       dotGz,
		MaxSize:     maxSize,
		Append:      append,
		rotateFunc:  func() bool { return false },
		MaxKeepDays: maxKeepDays,
	}

	if r.MaxSize > 0 {
		r.rotateFunc = func() bool { return r.curSize >= r.MaxSize }
	}

	return r
}

func (w *FileWriter) daysKeeping() {
	expired := time.Now().Add(time.Duration(w.MaxKeepDays) * -24 * time.Hour)
	matches, _ := filepath.Glob(matchExpiredFiles(w.FnTemplate, w.DotGz))
	for _, f := range matches {
		if stat, _ := os.Stat(f); stat != nil && stat.ModTime().Before(expired) {
			_ = os.Remove(f)
		}
	}
}

func matchExpiredFiles(fnTemplate, dotGz string) string {
	fn := timex.GlobName(fnTemplate)
	fn = filepath.Clean(fn)
	base, _, ext := SplitBaseIndexExt(fn)
	return base + "*" + ext + dotGz
}

func (w *FileWriter) Write(p []byte) (int, error) {
	timedFn := w.NewTimedFilename(w.FnTemplate, w.DotGz)

	for {
		fn := w.RotateFilename(timedFn)
		if fn == w.curFn {
			break
		}

		if ok, err := w.openFile(fn); err != nil {
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

func (w *gzipWriter) Close() error { return w.Writer.Close() }
func (w *gzipWriter) Flush() error { return w.Buf.Flush() }

type bufioWriter struct {
	*bufio.Writer
}

func (b *bufioWriter) Close() error { return b.Writer.Flush() }
func (b *bufioWriter) Flush() error { return b.Writer.Flush() }

func (w *FileWriter) openFile(fn string) (ok bool, err error) {
	_ = w.Close()
	if w.maxIndex == 2 { // rename bbb-2021-05-27-18-26.http to bbb-2021-05-27-18-26_00001.http
		_ = os.Rename(w.curFn+w.DotGz, SetFileIndex(w.curFn, 1)+w.DotGz)
	}

	w.file, err = os.OpenFile(fn+w.DotGz, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o660)
	if err != nil {
		return false, err
	}

	w.curFn = fn

	if w.DotGz != "" {
		gw := gzip.NewWriter(w.file)
		w.writer = &gzipWriter{Buf: bufio.NewWriter(gw), Writer: gw}
	} else {
		w.writer = &bufioWriter{Writer: bufio.NewWriter(w.file)}
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

		if w.MaxKeepDays > 0 {
			go w.daysKeeping()
		}
	}
	return nil
}

func (w *FileWriter) NewTimedFilename(template, dotGz string) string {
	fn := timex.FormatTime(time.Now(), template)
	fn = filepath.Clean(fn)

	if w.timedFn != fn {
		w.maxIndex = 1
		w.timedFn = fn
	}

	if w.curFn == "" { // 只有第一次检查最大文件索引号
		w.maxIndex, fn = FindMaxFileIndex(fn, dotGz)
	}

	return fn
}

func (w *FileWriter) RotateFilename(fn string) string {
	if w.rotateFunc() {
		w.maxIndex++
	}

	if w.maxIndex == 1 {
		return fn
	}
	return SetFileIndex(fn, w.maxIndex)
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

// FindMaxFileIndex finds the maxIndex index of a file like log-2021-05-27_00001.log.
func FindMaxFileIndex(path string, dotGz string) (int, string) {
	base, _, ext := SplitBaseIndexExt(path)
	matches, _ := filepath.Glob(base + "*" + ext + dotGz)
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
