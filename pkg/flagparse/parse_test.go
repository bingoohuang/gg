package flagparse

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

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
	ParseArgs(arg, []string{"app", "-i", "5003", "-out", "a", "-out", "b", "-my", "mymy", "-d", "10s"})
	assert.Equal(t, 10*time.Second, arg.Duration)
	assert.Equal(t, myFlag{Value: "mymy"}, arg.MyFlag)
	assert.Equal(t, []string{"a", "b"}, arg.Out)
	assert.Equal(t, 1234, arg.Port)
	assert.Equal(t, "5003", arg.Input)
	// ... use arg
}

// Arg is the application's argument options.
type Arg struct {
	Duration time.Duration `flag:"d"`
	MyFlag   myFlag        `flag:"my"`
	Out      []string      `flag:"out"`
	Port     int           `flag:"p" val:"1234"`
	Input    string        `flag:"i" val:"" required:"true"`
	Version  bool          `flag:"v" val:"false" usage:"Show version"`
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
