package extra

import (
	"strings"
	"unicode"

	"github.com/bingoohuang/gg/pkg/jsoni"
)

// SetNamingStrategy rename struct fields uniformly
func SetNamingStrategy(translate func(string) string) {
	jsoni.RegisterExtension(&NamingStrategyExtension{Translate: translate})
}

type NamingStrategyExtension struct {
	jsoni.DummyExtension
	Translate func(string) string
}

func (e *NamingStrategyExtension) UpdateStructDescriptor(sd *jsoni.StructDescriptor) {
	for _, f := range sd.Fields {
		if unicode.IsLower(rune(f.Field.Name()[0])) || f.Field.Name()[0] == '_' {
			continue
		}
		if tag, ok := f.Field.Tag().Lookup("json"); ok {
			tagParts := strings.Split(tag, ",")
			if tagParts[0] == "-" {
				continue // hidden field
			}
			if tagParts[0] != "" {
				continue // field explicitly named
			}
		}
		f.ToNames = []string{e.Translate(f.Field.Name())}
		f.FromNames = []string{e.Translate(f.Field.Name())}
	}
}

// LowerCaseWithUnderscores one strategy to SetNamingStrategy for. It will change HelloWorld to hello_world.
func LowerCaseWithUnderscores(name string) string {
	var newName []rune
	for i, c := range name {
		if i == 0 {
			newName = append(newName, unicode.ToLower(c))
		} else {
			if unicode.IsUpper(c) {
				newName = append(newName, '_')
				newName = append(newName, unicode.ToLower(c))
			} else {
				newName = append(newName, c)
			}
		}
	}
	return string(newName)
}
