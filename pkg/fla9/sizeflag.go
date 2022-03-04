package fla9

import (
	"github.com/bingoohuang/gg/pkg/man"
)

func NewSizeFlag(up *uint64, val string) Value {
	if val != "" {
		*up, _ = man.ParseBytes(val)
	}
	return &SizeFlag{Val: up}
}

type SizeFlag struct {
	Val *uint64
}

func (i *SizeFlag) String() string {
	if i.Val == nil {
		return "0"
	}
	return man.Bytes(*i.Val)
}

func (i *SizeFlag) Set(value string) (err error) {
	*i.Val, err = man.ParseBytes(value)
	return err
}

// SizeVar defines a size flag with specified name, default value, and usage string.
// The argument p points to an uint64 variable in which to store the value of the flag.
func SizeVar(p *uint64, name string, value string, usage string) {
	CommandLine.Var(NewSizeFlag(p, value), name, usage)
}

// Size defines a size flag with specified name, default value, and usage string.
// The return value is the address of an uint64 variable that stores the value of the flag.
func Size(name string, value string, usage string) *uint64 {
	return CommandLine.Size(name, value, usage)
}

// SizeVar defines an uint64 flag with specified name, default value, and usage string.
// The argument p points to an uint64 variable in which to store the value of the flag.
func (f *FlagSet) SizeVar(p *uint64, name string, value string, usage string) {
	f.Var(NewSizeFlag(p, value), name, usage)
}

// Size defines a size flag with specified name, default value, and usage string.
// The return value is the address of an uint64 variable that stores the value of the flag.
func (f *FlagSet) Size(name string, value string, usage string) *uint64 {
	p := new(uint64)
	f.SizeVar(p, name, value, usage)
	return p
}
