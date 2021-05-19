
```go
import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/flagparse"
)

func main() {
	arg := &Arg{}
	flagparse.Parse(arg)
	// ... use arg
}

// Arg is the application's argument options.
type Arg struct {
	Input   string `flag:"i" val:"" required:"true"`
	Version bool   `flag:"v" val:"false" usage:"Show version"`
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
```
