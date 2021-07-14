package flagparse

import (
	"github.com/bingoohuang/gg/pkg/fla9"
	"github.com/bingoohuang/gg/pkg/man"
)

func newSizeFlag(up *uint64, val string) fla9.Value {
	if val != "" {
		*up, _ = man.ParseBytes(val)
	}
	return &sizeFlag{up: up}
}

type sizeFlag struct {
	up *uint64
}

func (i *sizeFlag) String() string { return man.Bytes(*i.up) }

func (i *sizeFlag) Set(value string) (err error) {
	*i.up, err = man.ParseBytes(value)
	return err
}
