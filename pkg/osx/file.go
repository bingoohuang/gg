package osx

import (
	"github.com/bingoohuang/gg/pkg/gz"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// ReadFile reads a file content, if it's a .gz, decompress it.
func ReadFile(filename string) ([]byte, error) {
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
