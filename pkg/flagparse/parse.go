package flagparse

import (
	"embed"
	"fmt"
	"github.com/bingoohuang/gg/pkg/cast"
	"github.com/bingoohuang/gg/pkg/ctl"
	flag "github.com/bingoohuang/gg/pkg/fla9"
	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/bingoohuang/gg/pkg/v"
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
	initFiles            *embed.FS
}

type OptionsFn func(*Options)

func ProcessInit(initFiles *embed.FS) OptionsFn {
	return func(o *Options) {
		o.initFiles = initFiles
	}
}

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
	initing := false

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
			name = ss.ToLowerKebab(fi.Name)
		} else if strings.HasPrefix(name, ",") { // for shortName
			name = ss.ToLowerKebab(fi.Name) + name
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

		fullName, shortName := ss.Split2(name, ss.WithSeps(","))

		switch ft.Kind() {
		case reflect.Slice:
			switch ft.Elem().Kind() {
			case reflect.String:
				pp := p.(*[]string)
				f.Var(&ArrayFlags{pp: pp, Value: val}, name, usage)
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

			switch {
			case ss.AnyOf("pprof", fullName, shortName):
				pprof = pp
			case ss.AnyOf(options.flagName, fullName, shortName):
				options.cnf = pp
			}

		case reflect.Int:
			if count := t("count"); count == "true" {
				f.CountVar(p.(*int), name, cast.ToInt(val), usage)
			} else {
				f.IntVar(p.(*int), name, cast.ToInt(val), usage)
			}
		case reflect.Int32:
			f.Int32Var(p.(*int32), name, cast.ToInt32(val), usage)
		case reflect.Int64:
			f.Int64Var(p.(*int64), name, cast.ToInt64(val), usage)
		case reflect.Uint:
			f.UintVar(p.(*uint), name, cast.ToUint(val), usage)
		case reflect.Uint32:
			f.Uint32Var(p.(*uint32), name, cast.ToUint32(val), usage)
		case reflect.Uint64:
			if size == "true" {
				f.Var(flag.NewSizeFlag(p.(*uint64), val), name, usage)
			} else {
				f.Uint64Var(p.(*uint64), name, cast.ToUint64(val), usage)
			}
		case reflect.Bool:
			if fi.Name == "Init" {
				f.BoolVar(&initing, name, false, usage)
			} else {
				pp := p.(*bool)
				checkVersionShow = checkVersion(checkVersionShow, a, fi.Name, pp)
				f.BoolVar(pp, name, cast.ToBool(val), usage)
			}
		case reflect.Float32:
			f.Float32Var(p.(*float32), name, cast.ToFloat32(val), usage)
		case reflect.Float64:
			f.Float64Var(p.(*float64), name, cast.ToFloat64(val), usage)
		}
	}

	if u, ok := a.(UsageShower); ok {
		f.Usage = func() {
			fmt.Println(strings.TrimSpace(u.Usage()))
		}
	}

	if options.cnf != nil {
		fn, sn := ss.Split2(options.flagName, ss.WithSeps(","))
		if value, _ := FindFlag(args, fn, sn); value != "" || options.defaultCnf != "" {
			if err := LoadConfFile(value, options.defaultCnf, a); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(-1)
			}
		}
	}

	// 提前到这里，实际上是为了先解析出 --conf 参数，便于下面从配置文件载入数据
	// 但是，命令行应该优先级，应该比配置文件优先级高，为了解决这个矛盾
	// 需要把 --conf 参数置为第一个参数，并且使用自定义参数的形式，在解析到改参数时，
	// 立即从对应的配置文件加载所有配置，然后再依次处理其它命令行参数
	_ = f.Parse(args[1:])

	if checkVersionShow != nil {
		checkVersionShow()
	}
	if initing {
		ctl.Config{Initing: true, InitFiles: options.initFiles}.ProcessInit()
	}

	checkRequired(requiredVars, f)

	if pp, ok := a.(PostProcessor); ok {
		pp.PostProcess()
	}

	if pprof != nil && *pprof != "" {
		go startPprof(*pprof)
	}
}

func FindFlag(args []string, targetNames ...string) (value string, found bool) {
	for i := 1; i < len(args); i++ {
		s := args[i]
		if len(s) < 2 || s[0] != '-' {
			continue
		}
		numMinuses := 1
		if s[1] == '-' {
			numMinuses++
			if len(s) == 2 { // "--" terminates the flags
				break
			}
		}

		name := s[numMinuses:]
		if len(name) == 0 || name[0] == '-' || name[0] == '=' { // bad flag syntax: %s"
			continue
		}
		if strings.HasPrefix(name, "test.") { // ignore go test flags
			continue
		}

		// it's a flag. does it have an argument?
		hasValue := false
		for j := 1; j < len(name); j++ { // equals cannot be first
			if name[j] == '=' {
				value = name[j+1:]
				hasValue = true
				name = name[0:j]
				break
			}
		}

		if !ss.AnyOf(name, targetNames...) {
			continue
		}

		// It must have a value, which might be the next argument.
		if !hasValue && i+1 < len(args) {
			// value is the next arg
			hasValue = true
			value = args[i+1]
		}

		return value, true
	}

	return "", false
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
	for _, rv := range requiredVars {
		if rv.p != nil && *rv.p == "" || rv.pp != nil && len(*rv.pp) == 0 {
			requiredMissed++
			fmt.Printf("-%s is required\n", rv.name)
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
		} else {
			return func() {
				if *bp {
					fmt.Println(v.Version())
					os.Exit(0)
				}
			}
		}
	}

	return checker
}

type ArrayFlags struct {
	Value string
	pp    *[]string
}

func (i *ArrayFlags) String() string { return i.Value }

func (i *ArrayFlags) Set(value string) error {
	*i.pp = append(*i.pp, value)
	return nil
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
