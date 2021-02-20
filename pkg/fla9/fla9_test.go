package fla9_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/bingoohuang/gg/pkg/fla9"
)

func ExampleFlag() {
	var (
		name   string
		age    int
		length float64
		female bool
	)

	fla9.String("config", "testdata/test.conf", "help message")
	fla9.StringVar(&name, "name", "", "help message")
	fla9.IntVar(&age, "age", 0, "help message")
	fla9.Float64Var(&length, "length", 0, "help message")
	fla9.BoolVar(&female, "female", false, "help message")

	fla9.Parse()

	fmt.Println("length:", length)
	fmt.Println("age:", age)
	fmt.Println("name:", name)
	fmt.Println("female:", female)

	// Output:
	// length: 175.5
	// age: 2
	// name: Gloria
	// female: false
}

// Example 1: A single string flag called "species" with default value "gopher".
var species = fla9.String("species", "gopher", "the species we are studying")

// Example 2: Two flags sharing a variable, so we can have a shorthand.
// The order of initialization is undefined, so make sure both use the
// same default value. They must be set up with an init function.
var gopherType string

func init() {
	const (
		defaultGopher = "pocket"
		usage         = "the variety of gopher"
	)
	fla9.StringVar(&gopherType, "gopher_type", defaultGopher, usage)
	fla9.StringVar(&gopherType, "g", defaultGopher, usage+" (shorthand)")
}

// Example 3: A user-defined flag type, a slice of durations.
type interval []time.Duration

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (i *interval) String() string {
	return fmt.Sprint(*i)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (i *interval) Set(value string) error {
	// If we wanted to allow the flag to be set multiple times,
	// accumulating values, we would delete this if statement.
	// That would permit usages such as
	//	-deltaT 10s -deltaT 15s
	// and other combinations.
	if len(*i) > 0 {
		return errors.New("interval flag already set")
	}
	for _, dt := range strings.Split(value, ",") {
		duration, err := time.ParseDuration(dt)
		if err != nil {
			return err
		}
		*i = append(*i, duration)
	}
	return nil
}

// Define a flag to accumulate durations. Because it has a special type,
// we need to use the Var function and therefore create the flag during
// init.

var intervalFlag interval

func init() {
	// Tie the command-line flag to the intervalFlag variable and
	// set a usage message.
	fla9.Var(&intervalFlag, "deltaT", "comma-separated list of intervals to use between events")
}

func Example() {
	// All the interesting pieces are with the variables declared above, but
	// to enable the flag package to see the flags defined there, one must
	// execute, typically at the start of main (not init!):
	//	flag.Parse()
	// We don't run it here because this is not a main function and
	// the testing suite has already parsed the flags.
}

// Additional routines compiled into the package only during testing.

// ResetForTesting clears all flag state and sets the usage function as directed.
// After calling ResetForTesting, parse errors in flag handling will not
// exit the program.
func ResetForTesting(usage func()) {
	fla9.CommandLine = fla9.NewFlagSet(os.Args[0], fla9.ContinueOnError)
	fla9.Usage = usage
}

func boolString(s string) string {
	if s == "0" {
		return "false"
	}
	return "true"
}

func TestEverything(t *testing.T) {
	ResetForTesting(nil)
	fla9.Bool("test_bool", false, "bool value")
	fla9.Int("test_int", 0, "int value")
	fla9.Int64("test_int64", 0, "int64 value")
	fla9.Uint("test_uint", 0, "uint value")
	fla9.Uint64("test_uint64", 0, "uint64 value")
	fla9.String("test_string", "0", "string value")
	fla9.Float64("test_float64", 0, "float64 value")
	fla9.Duration("test_duration", 0, "time.Duration value")

	m := make(map[string]*fla9.Flag)
	desired := "0"
	visitor := func(f *fla9.Flag) {
		if len(f.Name) > 5 && f.Name[0:5] == "test_" {
			m[f.Name] = f
			ok := false
			switch {
			case f.Value.String() == desired:
				ok = true
			case f.Name == "test_bool" && f.Value.String() == boolString(desired):
				ok = true
			case f.Name == "test_duration" && f.Value.String() == desired+"s":
				ok = true
			}
			if !ok {
				t.Error("Visit: bad value", f.Value.String(), "for", f.Name)
			}
		}
	}
	fla9.VisitAll(visitor)
	if len(m) != 8 {
		t.Error("VisitAll misses some flags")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	m = make(map[string]*fla9.Flag)
	fla9.Visit(visitor)
	if len(m) != 0 {
		t.Errorf("Visit sees unset flags")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	// Now set all flags
	fla9.Set("test_bool", "true")
	fla9.Set("test_int", "1")
	fla9.Set("test_int64", "1")
	fla9.Set("test_uint", "1")
	fla9.Set("test_uint64", "1")
	fla9.Set("test_string", "1")
	fla9.Set("test_float64", "1")
	fla9.Set("test_duration", "1s")
	desired = "1"
	fla9.Visit(visitor)
	if len(m) != 8 {
		t.Error("Visit fails after set")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	// Now test they're visited in sort order.
	var flagNames []string
	fla9.Visit(func(f *fla9.Flag) { flagNames = append(flagNames, f.Name) })
	if !sort.StringsAreSorted(flagNames) {
		t.Errorf("flag names not sorted: %v", flagNames)
	}
}

func TestGet(t *testing.T) {
	ResetForTesting(nil)
	fla9.Bool("test_bool", true, "bool value")
	fla9.Int("test_int", 1, "int value")
	fla9.Int64("test_int64", 2, "int64 value")
	fla9.Uint("test_uint", 3, "uint value")
	fla9.Uint64("test_uint64", 4, "uint64 value")
	fla9.String("test_string", "5", "string value")
	fla9.Float64("test_float64", 6, "float64 value")
	fla9.Duration("test_duration", 7, "time.Duration value")

	visitor := func(f *fla9.Flag) {
		if len(f.Name) > 5 && f.Name[0:5] == "test_" {
			g, ok := f.Value.(fla9.Getter)
			if !ok {
				t.Errorf("Visit: value does not satisfy Getter: %T", f.Value)
				return
			}
			switch f.Name {
			case "test_bool":
				ok = g.Get() == true
			case "test_int":
				ok = g.Get() == int(1)
			case "test_int64":
				ok = g.Get() == int64(2)
			case "test_uint":
				ok = g.Get() == uint(3)
			case "test_uint64":
				ok = g.Get() == uint64(4)
			case "test_string":
				ok = g.Get() == "5"
			case "test_float64":
				ok = g.Get() == float64(6)
			case "test_duration":
				ok = g.Get() == time.Duration(7)
			}
			if !ok {
				t.Errorf("Visit: bad value %T(%v) for %s", g.Get(), g.Get(), f.Name)
			}
		}
	}
	fla9.VisitAll(visitor)
}

func TestUsage(t *testing.T) {
	called := false
	ResetForTesting(func() { called = true })
	if fla9.CommandLine.Parse([]string{"-x"}) == nil {
		t.Error("parse did not fail for unknown flag")
	}
	if !called {
		t.Error("did not call Usage for unknown flag")
	}
}

func testParse(f *fla9.FlagSet, t *testing.T) {
	if f.Parsed() {
		t.Error("f.Parse() = true before Parse")
	}
	boolFlag := f.Bool("bool", false, "bool value")
	bool2Flag := f.Bool("bool2", false, "bool2 value")
	intFlag := f.Int("int", 0, "int value")
	int64Flag := f.Int64("int64", 0, "int64 value")
	uintFlag := f.Uint("uint", 0, "uint value")
	uint64Flag := f.Uint64("uint64", 0, "uint64 value")
	stringFlag := f.String("string", "0", "string value")
	float64Flag := f.Float64("float64", 0, "float64 value")
	durationFlag := f.Duration("duration", 5*time.Second, "time.Duration value")
	extra := "one-extra-argument"
	args := []string{
		"-bool",
		"-bool2=true",
		"--int", "22",
		"--int64", "0x23",
		"-uint", "24",
		"--uint64", "25",
		"-string", "hello",
		"-float64", "2718e28",
		"-duration", "2m",
		extra,
	}
	if err := f.Parse(args); err != nil {
		t.Fatal(err)
	}
	if !f.Parsed() {
		t.Error("f.Parse() = false after Parse")
	}
	if *boolFlag != true {
		t.Error("bool flag should be true, is ", *boolFlag)
	}
	if *bool2Flag != true {
		t.Error("bool2 flag should be true, is ", *bool2Flag)
	}
	if *intFlag != 22 {
		t.Error("int flag should be 22, is ", *intFlag)
	}
	if *int64Flag != 0x23 {
		t.Error("int64 flag should be 0x23, is ", *int64Flag)
	}
	if *uintFlag != 24 {
		t.Error("uint flag should be 24, is ", *uintFlag)
	}
	if *uint64Flag != 25 {
		t.Error("uint64 flag should be 25, is ", *uint64Flag)
	}
	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is ", *stringFlag)
	}
	if *float64Flag != 2718e28 {
		t.Error("float64 flag should be 2718e28, is ", *float64Flag)
	}
	if *durationFlag != 2*time.Minute {
		t.Error("duration flag should be 2m, is ", *durationFlag)
	}
	if len(f.Args()) != 1 {
		t.Error("expected one argument, got", len(f.Args()))
	} else if f.Args()[0] != extra {
		t.Errorf("expected argument %q got %q", extra, f.Args()[0])
	}
}

func TestParse(t *testing.T) {
	ResetForTesting(func() { t.Error("bad parse") })
	testParse(fla9.CommandLine, t)
}

func TestFlagSetParse(t *testing.T) {
	testParse(fla9.NewFlagSet("test", fla9.ContinueOnError), t)
}

// Declare a user-defined flag type.
type flagVar []string

func (f *flagVar) String() string {
	return fmt.Sprint([]string(*f))
}

func (f *flagVar) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func TestUserDefined(t *testing.T) {
	var flags fla9.FlagSet
	flags.Init("test", fla9.ContinueOnError)
	var v flagVar
	flags.Var(&v, "v", "usage")
	if err := flags.Parse([]string{"-v", "1", "-v", "2", "-v=3"}); err != nil {
		t.Error(err)
	}
	if len(v) != 3 {
		t.Fatal("expected 3 args; got ", len(v))
	}
	expect := "[1 2 3]"
	if v.String() != expect {
		t.Errorf("expected value %q got %q", expect, v.String())
	}
}

func TestUserDefinedForCommandLine(t *testing.T) {
	const help = "HELP"
	var result string
	ResetForTesting(func() { result = help })
	fla9.Usage()
	if result != help {
		t.Fatalf("got %q; expected %q", result, help)
	}
}

// Declare a user-defined boolean flag type.
type boolFlagVar struct {
	count int
}

func (b *boolFlagVar) String() string {
	return fmt.Sprintf("%d", b.count)
}

func (b *boolFlagVar) Set(value string) error {
	if value == "true" {
		b.count++
	}
	return nil
}

func (b *boolFlagVar) IsBoolFlag() bool {
	return b.count < 4
}

func TestUserDefinedBool(t *testing.T) {
	var flags fla9.FlagSet
	flags.Init("test", fla9.ContinueOnError)
	var b boolFlagVar
	var err error
	flags.Var(&b, "b", "usage")
	if err = flags.Parse([]string{"-b", "-b", "-b", "-b=true", "-b=false", "-b", "barg", "-b"}); err != nil {
		if b.count < 4 {
			t.Error(err)
		}
	}

	if b.count != 4 {
		t.Errorf("want: %d; got: %d", 4, b.count)
	}

	if err == nil {
		t.Error("expected error; got none")
	}
}

func TestSetOutput(t *testing.T) {
	var flags fla9.FlagSet
	var buf bytes.Buffer
	flags.SetOutput(&buf)
	flags.Init("test", fla9.ContinueOnError)
	flags.Parse([]string{"-unknown"})
	if out := buf.String(); !strings.Contains(out, "-unknown") {
		t.Logf("expected output mentioning unknown; got %q", out)
	}
}

// This tests that one can reset the flags. This still works but not well, and is
// superseded by FlagSet.
func TestChangingArgs(t *testing.T) {
	ResetForTesting(func() { t.Fatal("bad parse") })
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-before", "subcmd", "-after", "args"}
	before := fla9.Bool("before", false, "")
	if err := fla9.CommandLine.Parse(os.Args[1:]); err != nil {
		t.Fatal(err)
	}
	cmd := fla9.Arg(0)
	os.Args = fla9.Args()
	after := fla9.Bool("after", false, "")
	fla9.Parse()
	args := fla9.Args()

	if !*before || cmd != "subcmd" || !*after || len(args) != 1 || args[0] != "args" {
		t.Fatalf("expected true subcmd true [args] got %v %v %v %v", *before, cmd, *after, args)
	}
}

// Test that -help invokes the usage message and returns ErrHelp.
func TestHelp(t *testing.T) {
	var helpCalled = false
	fs := fla9.NewFlagSet("help test", fla9.ContinueOnError)
	fs.Usage = func() { helpCalled = true }
	var fla bool
	fs.BoolVar(&fla, "fla", false, "regular fla")
	// Regular fla invocation should work
	err := fs.Parse([]string{"-fla=true"})
	if err != nil {
		t.Fatal("expected no error; got ", err)
	}
	if !fla {
		t.Error("fla was not set by -fla")
	}
	if helpCalled {
		t.Error("help called for regular fla")
		helpCalled = false // reset for next test
	}
	// Help fla should work as expected.
	err = fs.Parse([]string{"-help"})
	if err == nil {
		t.Fatal("error expected")
	}
	if err != fla9.ErrHelp {
		t.Fatal("expected ErrHelp; got ", err)
	}
	if !helpCalled {
		t.Fatal("help was not called")
	}
	// If we define a help fla, that should override.
	var help bool
	fs.BoolVar(&help, "help", false, "help fla")
	helpCalled = false
	err = fs.Parse([]string{"-help"})
	if err != nil {
		t.Fatal("expected no error for defined -help; got ", err)
	}
	if helpCalled {
		t.Fatal("help was called; should not have been for defined help fla")
	}
}

const defaultOutput = `  -A	for bootstrapping, allow 'any' type
  -Alongflagname
    	disable bounds checking
  -C	a boolean defaulting to true (default true)
  -D path
    	set relative path for local imports
  -F number
    	a non-zero number (default 2.7)
  -G float
    	a float that defaults to zero
  -N int
    	a non-zero int (default 27)
  -Z int
    	an int that defaults to zero
  -maxT timeout
    	set timeout for dial
`

func TestPrintDefaults(t *testing.T) {
	fs := fla9.NewFlagSet("print defaults test", fla9.ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.Bool("A", false, "for bootstrapping, allow 'any' type")
	fs.Bool("Alongflagname", false, "disable bounds checking")
	fs.Bool("C", true, "a boolean defaulting to true")
	fs.String("D", "", "set relative `path` for local imports")
	fs.Float64("F", 2.7, "a non-zero `number`")
	fs.Float64("G", 0, "a float that defaults to zero")
	fs.Int("N", 27, "a non-zero int")
	fs.Int("Z", 0, "an int that defaults to zero")
	fs.Duration("maxT", 0, "set `timeout` for dial")
	fs.PrintDefaults()
	got := buf.String()
	if got != defaultOutput {
		t.Errorf("got %q want %q\n", got, defaultOutput)
	}
}

// Test parsing a environment variables
func TestParseEnv(t *testing.T) {

	syscall.Setenv("BOOL", "")
	syscall.Setenv("BOOL2", "true")
	syscall.Setenv("INT", "22")
	syscall.Setenv("INT64", "0x23")
	syscall.Setenv("UINT", "24")
	syscall.Setenv("UINT64", "25")
	syscall.Setenv("STRING", "hello")
	syscall.Setenv("FLOAT64", "2718e28")
	syscall.Setenv("DURATION", "2m")

	f := fla9.NewFlagSet(os.Args[0], fla9.ContinueOnError)

	boolFlag := f.Bool("bool", false, "bool value")
	bool2Flag := f.Bool("bool2", false, "bool2 value")
	intFlag := f.Int("int", 0, "int value")
	int64Flag := f.Int64("int64", 0, "int64 value")
	uintFlag := f.Uint("uint", 0, "uint value")
	uint64Flag := f.Uint64("uint64", 0, "uint64 value")
	stringFlag := f.String("string", "0", "string value")
	float64Flag := f.Float64("float64", 0, "float64 value")
	durationFlag := f.Duration("duration", 5*time.Second, "time.Duration value")

	err := f.ParseEnv(os.Environ())
	if err != nil {
		t.Fatal("expected no error; got ", err)
	}
	if *boolFlag != true {
		t.Error("bool flag should be true, is ", *boolFlag)
	}
	if *bool2Flag != true {
		t.Error("bool2 flag should be true, is ", *bool2Flag)
	}
	if *intFlag != 22 {
		t.Error("int flag should be 22, is ", *intFlag)
	}
	if *int64Flag != 0x23 {
		t.Error("int64 flag should be 0x23, is ", *int64Flag)
	}
	if *uintFlag != 24 {
		t.Error("uint flag should be 24, is ", *uintFlag)
	}
	if *uint64Flag != 25 {
		t.Error("uint64 flag should be 25, is ", *uint64Flag)
	}
	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is ", *stringFlag)
	}
	if *float64Flag != 2718e28 {
		t.Error("float64 flag should be 2718e28, is ", *float64Flag)
	}
	if *durationFlag != 2*time.Minute {
		t.Error("duration flag should be 2m, is ", *durationFlag)
	}
}

// Test parsing a configuration file
func TestParseFile(t *testing.T) {
	f := fla9.NewFlagSet(os.Args[0], fla9.ContinueOnError)

	boolFlag := f.Bool("bool", false, "bool value")
	bool2Flag := f.Bool("bool2", false, "bool2 value")
	intFlag := f.Int("int", 0, "int value")
	int64Flag := f.Int64("int64", 0, "int64 value")
	uintFlag := f.Uint("uint", 0, "uint value")
	uint64Flag := f.Uint64("uint64", 0, "uint64 value")
	stringFlag := f.String("string", "0", "string value")
	float64Flag := f.Float64("float64", 0, "float64 value")
	durationFlag := f.Duration("duration", 5*time.Second, "time.Duration value")

	err := f.ParseFile("./testdata/test.conf", true)
	if err != nil {
		t.Fatal("expected no error; got ", err)
	}
	if *boolFlag != true {
		t.Error("bool flag should be true, is ", *boolFlag)
	}
	if *bool2Flag != true {
		t.Error("bool2 flag should be true, is ", *bool2Flag)
	}
	if *intFlag != 22 {
		t.Error("int flag should be 22, is ", *intFlag)
	}
	if *int64Flag != 0x23 {
		t.Error("int64 flag should be 0x23, is ", *int64Flag)
	}
	if *uintFlag != 24 {
		t.Error("uint flag should be 24, is ", *uintFlag)
	}
	if *uint64Flag != 25 {
		t.Error("uint64 flag should be 25, is ", *uint64Flag)
	}
	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is ", *stringFlag)
	}
	if *float64Flag != 2718e28 {
		t.Error("float64 flag should be 2718e28, is ", *float64Flag)
	}
	if *durationFlag != 2*time.Minute {
		t.Error("duration flag should be 2m, is ", *durationFlag)
	}
}

func TestParseFileUnknownFlag(t *testing.T) {
	f := fla9.NewFlagSet("test", fla9.ContinueOnError)
	if err := f.ParseFile("testdata/bad_test.conf", false); err == nil {
		t.Error("parse did not fail for unknown flag; ", err)
	}
}

func TestDefaultConfigFlagname(t *testing.T) {
	f := fla9.NewFlagSet("test", fla9.ContinueOnError)

	f.Bool("bool", false, "bool value")
	f.Bool("bool2", false, "bool2 value")
	f.Int("int", 0, "int value")
	f.Int64("int64", 0, "int64 value")
	f.Uint("uint", 0, "uint value")
	f.Uint64("uint64", 0, "uint64 value")
	stringFlag := f.String("string", "0", "string value")
	f.Float64("float64", 0, "float64 value")
	f.Duration("duration", 5*time.Second, "time.Duration value")

	f.String(fla9.DefaultConfigFlagName, "./testdata/test.conf", "config path")

	if err := os.Unsetenv("STRING"); err != nil {
		t.Error(err)
	}

	if err := f.Parse([]string{}); err != nil {
		t.Error("parse failed; ", err)
	}

	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is", *stringFlag)
	}
}

func TestDefaultConfigFlagnameMissingFile(t *testing.T) {
	f := fla9.NewFlagSet("test", fla9.ContinueOnError)
	f.String(fla9.DefaultConfigFlagName, "./testdata/missing", "config path")

	if err := os.Unsetenv("STRING"); err != nil {
		t.Error(err)
	}
	if err := f.Parse([]string{}); err == nil {
		t.Error("expected error of missing config file, got nil")
	}
}

func TestFlagSetParseErrors(t *testing.T) {
	fs := fla9.NewFlagSet("test", fla9.ContinueOnError)
	fs.Int("int", 0, "int value")

	args := []string{"-int", "bad"}
	expected := `invalid value "bad" for flag -int: strconv.ParseInt: parsing "bad": invalid syntax`
	if err := fs.Parse(args); err == nil || err.Error() != expected {
		t.Errorf("expected error %q parsing from args, got: %v", expected, err)
	}

	if err := os.Setenv("INT", "bad"); err != nil {
		t.Fatalf("error setting env: %s", err.Error())
	}
	expected = `invalid value "bad" for environment variable int: strconv.ParseInt: parsing "bad": invalid syntax`
	if err := fs.Parse([]string{}); err == nil || err.Error() != expected {
		t.Errorf("expected error %q parsing from env, got: %v", expected, err)
	}
	if err := os.Unsetenv("INT"); err != nil {
		t.Fatalf("error unsetting env: %s", err.Error())
	}

	fs.String("config", "", "config filename")
	args = []string{"-config", "testdata/bad_test.conf"}
	expected = `invalid value "bad" for configuration variable int: strconv.ParseInt: parsing "bad": invalid syntax`
	if err := fs.Parse(args); err == nil || err.Error() != expected {
		t.Errorf("expected error %q parsing from config, got: %v", expected, err)
	}
}

func TestTestingPackageFlags(t *testing.T) {
	f := fla9.NewFlagSet("test", fla9.ContinueOnError)
	if err := f.Parse([]string{"-test.v", "-test.count", "1"}); err != nil {
		t.Error(err)
	}
}
