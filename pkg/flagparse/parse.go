package flagparse

import (
	"flag"
	"fmt"
	"github.com/bingoohuang/gg/pkg/cast"
	"os"
	"reflect"
	"strings"
)

type VersionShower interface {
	VersionInfo() string
}

type UsageShower interface {
	Usage() string
}

type requiredVar struct {
	name  string
	value *string
}

func Parse(a interface{}) {
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	checkVersionShow := func() {}
	requiredVars := make([]requiredVar, 0)

	ra := reflect.ValueOf(a).Elem()
	t := ra.Type()
	for i := 0; i < t.NumField(); i++ {
		fi, fv := t.Field(i), ra.Field(i)
		tag := fi.Tag.Get
		name := tag("flag")
		if name == "" || !fv.CanAddr() {
			continue
		}

		val, usage, required := tag("val"), tag("usage"), tag("required")
		p := fv.Addr().Interface()
		switch fi.Type.Kind() {
		case reflect.String:
			pp := p.(*string)
			f.StringVar(pp, name, val, usage)
			if required == "true" {
				requiredVars = append(requiredVars, requiredVar{name: name, value: pp})
			}
		case reflect.Int:
			f.IntVar(p.(*int), name, cast.ToInt(val), usage)
		case reflect.Bool:
			bp := p.(*bool)
			checkVersionShow = checkVersion(a, fi.Name, bp)
			f.BoolVar(bp, name, cast.ToBool(val), usage)
		}
	}

	if usageShower, ok := a.(UsageShower); ok {
		f.Usage = func() {
			fmt.Println(strings.TrimSpace(usageShower.Usage()))
		}
	}

	_ = f.Parse(os.Args[1:])

	checkVersionShow()
	checkRequired(requiredVars, f)
}

func checkRequired(requiredVars []requiredVar, f *flag.FlagSet) {
	requiredMissed := 0
	for _, v := range requiredVars {
		if *v.value == "" {
			requiredMissed++
			fmt.Printf("-%s is required\n", v.name)
		}
	}

	if requiredMissed > 0 {
		f.Usage()
		os.Exit(1)
	}
}

func checkVersion(arg interface{}, fiName string, bp *bool) func() {
	if fiName == "Version" {
		if vs, ok := arg.(VersionShower); ok {
			return func() {
				if *bp {
					fmt.Println(vs.VersionInfo())
					os.Exit(0)
				}
			}
		}
	}

	return func() {}
}
