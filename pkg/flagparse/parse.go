package flagparse

import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/cast"
	flag "github.com/bingoohuang/gg/pkg/fla9"
	"log"
	"net/http"
	_ "net/http/pprof" // Comment this line to disable pprof endpoint.
	"os"
	"reflect"
	"strings"
	"time"
)

type PostProcessor interface {
	PostProcess()
}

type VersionShower interface {
	VersionInfo() string
}

type UsageShower interface {
	Usage() string
}

type requiredVar struct {
	name string
	p    *string
	pp   *[]string
}

type Options struct {
	flagName, defaultCnf string
	cnf                  *string
}

type OptionsFn func(*Options)

func AutoLoadYaml(flagName, defaultCnf string) OptionsFn {
	return func(o *Options) {
		o.flagName = flagName
		o.defaultCnf = defaultCnf
	}
}

func Parse(a interface{}, optionFns ...OptionsFn) {
	ParseArgs(a, os.Args, optionFns...)
}

func ParseArgs(a interface{}, args []string, optionFns ...OptionsFn) {
	options := createOptions(optionFns)

	f := flag.NewFlagSet(args[0], flag.ExitOnError)
	var checkVersionShow func()
	requiredVars := make([]requiredVar, 0)

	var pprof *string

	ra := reflect.ValueOf(a).Elem()
	rt := ra.Type()
	for i := 0; i < rt.NumField(); i++ {
		fi, fv := rt.Field(i), ra.Field(i)
		if fi.PkgPath != "" { // ignore unexported
			continue
		}

		t := fi.Tag.Get
		name := t("flag")
		if name == "-" || !fv.CanAddr() {
			continue
		}

		if name == "" {
			name = toFlagName(fi.Name)
		}

		val, usage, required, size := t("val"), t("usage"), t("required"), t("size")
		p := fv.Addr().Interface()
		ft := fi.Type
		if reflect.PtrTo(ft).Implements(flagValueType) {
			f.Var(p.(flag.Value), name, usage)
			continue
		} else if ft == timeDurationType {
			f.DurationVar(p.(*time.Duration), name, cast.ToDuration(val), usage)
			continue
		}

		switch ft.Kind() {
		case reflect.Slice:
			switch ft.Elem().Kind() {
			case reflect.String:
				pp := p.(*[]string)
				f.Var(&arrayFlags{pp: pp, Value: val}, name, usage)
				if required == "true" {
					requiredVars = append(requiredVars, requiredVar{name: name, pp: pp})
				}
			}
		case reflect.String:
			pp := p.(*string)
			f.StringVar(pp, name, val, usage)
			if required == "true" {
				requiredVars = append(requiredVars, requiredVar{name: name, p: pp})
			}

			switch name {
			case "pprof":
				pprof = pp
			case options.flagName:
				options.cnf = pp
			}

		case reflect.Int:
			if count := t("count"); count == "true" {
				f.CountVar(p.(*int), name, cast.ToInt(val), usage)
			} else {
				f.IntVar(p.(*int), name, cast.ToInt(val), usage)
			}
		case reflect.Bool:
			pp := p.(*bool)
			checkVersionShow = checkVersion(checkVersionShow, a, fi.Name, pp)
			f.BoolVar(pp, name, cast.ToBool(val), usage)
		case reflect.Float64:
			f.Float64Var(p.(*float64), name, cast.ToFloat64(val), usage)
		case reflect.Uint64:
			if size == "true" {
				f.Var(newSizeFlag(p.(*uint64), val), name, usage)
			}
		case reflect.Int64:
			f.Int64Var(p.(*int64), name, cast.ToInt64(val), usage)
		case reflect.Uint:
			f.UintVar(p.(*uint), name, cast.ToUint(val), usage)
		}
	}

	if v, ok := a.(UsageShower); ok {
		f.Usage = func() {
			fmt.Println(strings.TrimSpace(v.Usage()))
		}
	}

	_ = f.Parse(args[1:])
	if checkVersionShow != nil {
		checkVersionShow()
	}

	if options.cnf != nil {
		if err := LoadConfFile(*options.cnf, options.defaultCnf, a); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(0)
		}
	}

	checkRequired(requiredVars, f)

	if v, ok := a.(PostProcessor); ok {
		v.PostProcess()
	}

	if pprof != nil && *pprof != "" {
		go startPprof(*pprof)
	}

}

func createOptions(fns []OptionsFn) *Options {
	options := &Options{}
	for _, f := range fns {
		f(options)
	}

	return options
}

var (
	timeDurationType = reflect.TypeOf(time.Duration(0))
	flagValueType    = reflect.TypeOf((*flag.Value)(nil)).Elem()
)

func checkRequired(requiredVars []requiredVar, f *flag.FlagSet) {
	requiredMissed := 0
	for _, v := range requiredVars {
		if v.p != nil && *v.p == "" || v.pp != nil && len(*v.pp) == 0 {
			requiredMissed++
			fmt.Printf("-%s is required\n", v.name)
		}
	}

	if requiredMissed > 0 {
		f.Usage()
		os.Exit(1)
	}
}

func checkVersion(checker func(), arg interface{}, fiName string, bp *bool) func() {
	if checker == nil && fiName == "Version" {
		if vs, ok := arg.(VersionShower); ok {
			return func() {
				if *bp {
					fmt.Println(vs.VersionInfo())
					os.Exit(0)
				}
			}
		}
	}

	return checker
}

type arrayFlags struct {
	Value string
	pp    *[]string
}

func (i *arrayFlags) String() string { return i.Value }

func (i *arrayFlags) Set(value string) error {
	*i.pp = append(*i.pp, value)
	return nil
}

func toFlagName(name string) string {
	var sb strings.Builder

	isUpper := func(c uint8) bool { return 'A' <= c && c <= 'Z' }

	for i := 0; i < len(name); i++ {
		c := name[i]
		if isUpper(c) {
			if sb.Len() > 0 {
				if i+1 < len(name) && (!(i-1 >= 0 && isUpper(name[i-1])) || !isUpper(name[i+1])) {
					sb.WriteByte('-')
				}
			}
			sb.WriteByte(c - 'A' + 'a')
		} else {
			sb.WriteByte(c)
		}
	}

	return sb.String()
}

func startPprof(pprofAddr string) {
	pprofHostPort := pprofAddr
	parts := strings.Split(pprofHostPort, ":")
	if len(parts) == 2 && parts[0] == "" {
		pprofHostPort = fmt.Sprintf("localhost:%s", parts[1])
	}

	log.Printf("I! Starting pprof HTTP server at: http://%s/debug/pprof", pprofHostPort)
	if err := http.ListenAndServe(pprofAddr, nil); err != nil {
		log.Fatal("E! " + err.Error())
	}
}
