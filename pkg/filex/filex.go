package filex

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/bingoohuang/gg/pkg/iox"
	"io"
	"log"
	"os"
	"strings"
)

// LinesChan read file into lines.
func LinesChan(filePath string, chSize int) (ch chan string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(f)
	s.Split(ScanLines)
	ch = make(chan string, chSize)
	go func() {
		defer iox.Close(f)
		defer close(ch)

		for s.Scan() {
			t := s.Text()
			t = strings.TrimSpace(t)
			if len(t) > 0 {
				ch <- t
			}
		}

		if err := s.Err(); err != nil {
			log.Printf("E! scan file %s lines  error: %v", filePath, err)
		}
	}()

	return ch, nil
}

// ScanLines is a split function for a Scanner that returns each line of
// text, with end-of-line marker. The returned line may
// be empty. The end-of-line marker is one optional carriage return followed
// by one mandatory newline. In regular expression notation, it is `\r?\n`.
// The last non-empty line of input will be returned even if it has no
// newline.
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// Lines read file into lines.
func Lines(filePath string) (lines []string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	return lines, s.Err()
}

// Open opens file successfully or panic.
func Open(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}

	return r
}

type AppendOptions struct {
	BackOffset int64
}

type AppendOptionsFn func(*AppendOptions)

func WithBackOffset(backOffset int64) AppendOptionsFn {
	return func(o *AppendOptions) {
		o.BackOffset = backOffset
	}
}

func Append(name string, data []byte, options ...AppendOptionsFn) (int, error) {
	option := &AppendOptions{}
	for _, fn := range options {
		fn(option)
	}

	// If the file doesn't exist, create it, or append to the file

	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	if _, err := f.Seek(option.BackOffset, io.SeekEnd); err != nil {
		return 0, err
	}

	n, err := f.Write(data)
	if err != nil {
		return n, err
	}

	return n, nil
}

func Exists(name string) bool {
	ok, _ := ExistsErr(name)
	return ok
}

func ExistsErr(name string) (bool, error) {
	if _, err := os.Stat(name); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return false, err
	}
}
