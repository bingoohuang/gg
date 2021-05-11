package setstruct

import (
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Set set struct fields by valueGenerator.
func Set(dstStructPtr interface{}, valueGenerator func(reflect.StructField) string) {
	vv := reflect.ValueOf(dstStructPtr).Elem()
	vt := vv.Type()
	for i := 0; i < vv.NumField(); i++ {
		f := vv.Field(i)
		if !f.CanSet() {
			continue
		}

		if v := valueGenerator(vt.Field(i)); v != "" {
			SetField(f, v)
		}
	}
}

// SetField set struct field by its string representation.
func SetField(f reflect.Value, v string) {
	t := f.Type()
	if reflect.PtrTo(t).Implements(UpdaterType) {
		if err := f.Addr().Interface().(Updater).ValueOf(v); err != nil {
			log.Printf("W! failed to parse %s as %s, error:%v", v, t, err)
		}
		return
	}

	for typ, parser := range StringParserMap {
		if t.AssignableTo(typ) {
			if vv, err := parser(v); err != nil {
				log.Printf("W! failed to parse %s as %s, error:%v", v, typ, err)
			} else {
				f.Set(vv.Convert(t))
			}
			return
		}
	}

	switch f.Kind() {
	case reflect.String:
		f.Set(reflect.ValueOf(v))
	case reflect.Bool:
		f.SetBool(ParseBool(v))
	case reflect.Int,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint64:
		if j, ok := ParseInt64(v); ok {
			f.SetInt(j)
		}
	case reflect.Float32, reflect.Float64:
		if j, ok := ParseFloat64(v); ok {
			f.SetFloat(j)
		}
	}
}

// StringParserMap is the map from reflect type to a string parser.
var StringParserMap = map[reflect.Type]StringParser{}

// RegisterStringParser registers a string parser to a reflect type.
func RegisterStringParser(typ reflect.Type, parser StringParser) {
	StringParserMap[typ] = parser
}

func init() {
	RegisterStringParser(reflect.TypeOf(time.Duration(0)), func(s string) (reflect.Value, error) {
		if d, err := time.ParseDuration(s); err != nil {
			return reflect.Value{}, err
		} else {
			return reflect.ValueOf(d), nil
		}
	})
	RegisterStringParser(reflect.TypeOf([]string{}), func(s string) (reflect.Value, error) {
		ss := strings.Split(s, ",")
		sd := make([]string, 0, len(ss))
		for _, si := range ss {
			if si = strings.TrimSpace(si); si != "" {
				sd = append(sd, si)
			}
		}

		return reflect.ValueOf(sd), nil
	})
	RegisterStringParser(reflect.TypeOf(map[string]string{}), func(s string) (reflect.Value, error) {
		return reflect.ValueOf(ParseStringToStringMap(s, ",", "=")), nil
	})
}

// StringParser is the func prototype to parse a string s to a reflect value.
type StringParser func(s string) (reflect.Value, error)

// Updater is the interface for a struct pointer receiver to be assigned a value by string s.
type Updater interface {
	// ValueOf assigned by string s to the pointer receiver.
	ValueOf(s string) error
}

// UpdaterType is the reflect type of Assigner interface.
var UpdaterType = reflect.TypeOf((*Updater)(nil)).Elem()

// ParseStringToStringMap parses a string to string map from a string.
func ParseStringToStringMap(val string, kkSep, kvSep string) map[string]string {
	m := make(map[string]string)
	for _, pair := range strings.Split(val, kkSep) {
		kv := strings.SplitN(pair, kvSep, 2)
		k := strings.TrimSpace(kv[0])
		if k == "" {
			continue
		}
		if len(kv) == 2 {
			m[k] = strings.TrimSpace(kv[1])
		} else {
			m[k] = ""
		}
	}
	return m
}

// ParseFloat64 parses string s as float64.
func ParseFloat64(s string) (float64, bool) {
	if j, err := strconv.ParseFloat(s, 64); err == nil {
		return j, true
	} else {
		log.Printf("W! failed to parse float64 %s, error: %v", s, err)
		return 0, false
	}
}

// ParseInt64 parses string s as int64.
func ParseInt64(s string) (int64, bool) {
	if j, err := strconv.ParseInt(s, 10, 64); err == nil {
		return j, true
	} else {
		log.Printf("W! failed to parse int64 %s, error: %v", s, err)
		return 0, false
	}
}

// ParseBool parses string s as a bool.
func ParseBool(s string) bool {
	switch u := strings.ToLower(s); u {
	case "1", "true", "t", "on", "yes", "y":
		return true
	}

	return false
}
