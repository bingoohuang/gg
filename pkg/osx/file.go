package osx

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bingoohuang/gg/pkg/gz"
)

// ReadFile reads a file content, if it's a .gz, decompress it.
func ReadFile(filename string) []byte {
	data, err := ReadFileE(filename)
	if err != nil {
		log.Fatalf("read file %s failed: %v", filename, err)
	}
	return data
}

// ReadFileE reads a file content, if it's a .gz, decompress it.
func ReadFileE(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(filename, ".gz") {
		data, err = gz.Ungzip(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func Remove(f string) {
	if err := os.Remove(f); err != nil {
		log.Printf("E! remove %s failed: %v", f, err)
	}
}
