package flagparse

import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/cast"
	flag "github.com/bingoohuang/gg/pkg/fla9"
	"os"
	"reflect"
	"strings"
	"time"
)

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

func Parse(a interface{}) {
	ParseArgs(a, os.Args)
}

func ParseArgs(a interface{}, args []string) {
	f := flag.NewFlagSet(args[0], flag.ExitOnError)
	checkVersionShow := func() {}
	checkVersionChecked := false
	requiredVars := make([]requiredVar, 0)

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

		val, usage, required := t("val"), t("usage"), t("required")
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
		case reflect.Int:
			f.IntVar(p.(*int), name, cast.ToInt(val), usage)
		case reflect.Bool:
			pp := p.(*bool)
			checkVersionShow, checkVersionChecked = checkVersion(checkVersionChecked, a, fi.Name, pp)
			f.BoolVar(pp, name, cast.ToBool(val), usage)
		case reflect.Float64:
			f.Float64Var(p.(*float64), name, cast.ToFloat64(val), usage)
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
	checkVersionShow()
	checkRequired(requiredVars, f)
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

func checkVersion(checked bool, arg interface{}, fiName string, bp *bool) (func(), bool) {
	if !checked && fiName == "Version" {
		if vs, ok := arg.(VersionShower); ok {
			return func() {
				if *bp {
					fmt.Println(vs.VersionInfo())
					os.Exit(0)
				}
			}, true
		}
	}

	return func() {}, false
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
