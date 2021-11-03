package filex

import (
	"bufio"
	"errors"
	"io"
	"os"
)

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
