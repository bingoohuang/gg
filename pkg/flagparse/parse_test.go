package flagparse

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// Arg is the application's argument options.
type Arg struct {
	Duration time.Duration `flag:"d"`
	MyFlag   myFlag        `flag:"my"`
	Out      []string
	Port     int    `flag:"p" val:"1234"`
	Input    string `flag:"i" val:"" required:"true"`
	Version  bool   `val:"false" usage:"Show version"`
	Other    string `flag:"-"`
	V        int    `flag:"v" count:"true"`
	Size     uint64 `size:"true" val:"10MiB"`
	Pmem     float32
}

type myFlag struct {
	Value string
}

func (i *myFlag) String() string { return i.Value }

func (i *myFlag) Set(value string) error {
	i.Value = value
	return nil
}

func TestParse(t *testing.T) {
	arg := &Arg{}
	ParseArgs(arg, []string{"app", "-i", "5003", "-out", "a", "-out", "b", "-my", "mymy", "-d", "10s", "-vvv", "-size", "2KiB", "-pmem", "0.618"})
	assert.Equal(t, 10*time.Second, arg.Duration)
	assert.Equal(t, myFlag{Value: "mymy"}, arg.MyFlag)
	assert.Equal(t, []string{"a", "b"}, arg.Out)
	assert.Equal(t, 1234, arg.Port)
	assert.Equal(t, "5003", arg.Input)
	assert.Equal(t, 3, arg.V)
	assert.Equal(t, uint64(2*1024), arg.Size)
	assert.Equal(t, float32(0.618), arg.Pmem)
	// ... use arg
}

// Usage is optional for customized show.
func (a Arg) Usage() string {
	return fmt.Sprintf(`
Usage of pcap (%s):
  -i string HTTP port to capture, or BPF, or pcap file
  -v        Show version
`, a.VersionInfo())
}

// VersionInfo is optional for customized version.
func (a Arg) VersionInfo() string { return "v0.0.2 2021-05-19 08:33:18" }

func TestTagName(t *testing.T) {
	assert.Equal(t, "abc", toFlagName("ABC"))
	assert.Equal(t, "hello-world", toFlagName("HelloWorld"))
	assert.Equal(t, "hello-url", toFlagName("HelloURL"))
	assert.Equal(t, "hello-url-addr", toFlagName("HelloURLAddr"))
}
