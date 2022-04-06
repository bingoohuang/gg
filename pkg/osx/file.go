package osx

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bingoohuang/gg/pkg/gz"
)

func Remove(f string) {
	if err := os.Remove(f); err != nil {
		log.Printf("E! remove %s failed: %v", f, err)
	}
}

type ReadFileConfig struct {
	AutoUncompress bool
	FatalOnError   bool
}

type ReadFileResult struct {
	Data []byte
	Err  error
}

// ReadFile reads a file content, if it's a .gz, decompress it.
func ReadFile(filename string, fns ...ReadFileConfigFn) (rr ReadFileResult) {
	config := (ReadFileConfigFns(fns)).Create()

	defer func() {
		if config.FatalOnError && rr.Err != nil {
			log.Fatal(rr.Err)
		}
	}()

	data, err := os.ReadFile(filename)
	if err != nil {
		rr.Err = fmt.Errorf("read file %s failed: %w", filename, err)
		return rr
	}

	if config.AutoUncompress && strings.HasSuffix(filename, ".gz") {
		if data, err = gz.Ungzip(data); err != nil {
			rr.Err = fmt.Errorf("Ungzip file %s failed: %w", filename, err)
			return rr
		}
	}

	return ReadFileResult{Data: data}
}

type ReadFileConfigFn func(*ReadFileConfig)

func WithAutoUncompress(v bool) ReadFileConfigFn {
	return func(config *ReadFileConfig) {
		config.AutoUncompress = v
	}
}

type ReadFileConfigFns []ReadFileConfigFn

func (fns ReadFileConfigFns) Create() (config ReadFileConfig) {
	for _, fn := range fns {
		fn(&config)
	}

	return config
}
