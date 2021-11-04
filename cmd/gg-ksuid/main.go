package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/bingoohuang/gg/pkg/uid"
)

var (
	count   int
	format  string
	tpltxt  string
	verbose bool
)

func init() {
	flag.IntVar(&count, "n", 1, "Number of KSUIDs to generate when called with no other arguments.")
	flag.StringVar(&format, "f", "string", "One of string, inspect, time, timestamp, payload, raw, or template.")
	flag.StringVar(&tpltxt, "t", "", "The Go template used to format the output.")
	flag.BoolVar(&verbose, "v", false, "Turn on verbose mode.")
}

func main() {
	flag.Parse()
	args := flag.Args()

	var print func(uid.KSUID)
	switch format {
	case "string":
		print = func(id uid.KSUID) { fmt.Println(id.String()) }
	case "inspect":
		print = printInspect
	case "time":
		print = func(id uid.KSUID) { fmt.Println(id.Time()) }
	case "timestamp":
		print = func(id uid.KSUID) { fmt.Println(id.Timestamp()) }
	case "payload":
		print = func(id uid.KSUID) { os.Stdout.Write(id.Payload()) }
	case "raw":
		print = func(id uid.KSUID) { os.Stdout.Write(id.Bytes()) }
	case "template":
		print = printTemplate
	default:
		fmt.Println("Bad formatting function:", format)
		os.Exit(1)
	}

	if len(args) == 0 {
		for i := 0; i < count; i++ {
			args = append(args, uid.New().String())
		}
	}

	var ids []uid.KSUID
	for _, arg := range args {
		id, err := uid.Parse(arg)
		if err != nil {
			fmt.Printf("Error when parsing %q: %s\n\n", arg, err)
			flag.PrintDefaults()
			os.Exit(1)
		}
		ids = append(ids, id)
	}

	for _, id := range ids {
		if verbose {
			fmt.Printf("%s: ", id)
		}
		print(id)
	}
}

func printInspect(id uid.KSUID) {
	const inspectFormat = `
REPRESENTATION:

  String: %v
     Raw: %v

COMPONENTS:

       Time: %v
  Timestamp: %v
    Payload: %v

`
	fmt.Printf(inspectFormat,
		id.String(),
		strings.ToUpper(hex.EncodeToString(id.Bytes())),
		id.Time(),
		id.Timestamp(),
		strings.ToUpper(hex.EncodeToString(id.Payload())),
	)
}

func printTemplate(id uid.KSUID) {
	b := &bytes.Buffer{}
	t := template.Must(template.New("").Parse(tpltxt))
	t.Execute(b, struct {
		String    string
		Raw       string
		Time      time.Time
		Timestamp uint32
		Payload   string
	}{
		String:    id.String(),
		Raw:       strings.ToUpper(hex.EncodeToString(id.Bytes())),
		Time:      id.Time(),
		Timestamp: id.Timestamp(),
		Payload:   strings.ToUpper(hex.EncodeToString(id.Payload())),
	})
	b.WriteByte('\n')
	io.Copy(os.Stdout, b)
}
