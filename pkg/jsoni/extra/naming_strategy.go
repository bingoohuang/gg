package extra

import (
	"github.com/bingoohuang/gg/pkg/jsoni"
	"strings"
	"unicode"
)

// SetNamingStrategy rename struct fields uniformly
func SetNamingStrategy(translate func(string) string) {
	jsoni.RegisterExtension(&namingStrategyExtension{jsoni.DummyExtension{}, translate})
}

type namingStrategyExtension struct {
	jsoni.DummyExtension
	translate func(string) string
}

func (e *namingStrategyExtension) UpdateStructDescriptor(structDescriptor *jsoni.StructDescriptor) {
	for _, binding := range structDescriptor.Fields {
		if unicode.IsLower(rune(binding.Field.Name()[0])) || binding.Field.Name()[0] == '_' {
			continue
		}
		tag, hastag := binding.Field.Tag().Lookup("json")
		if hastag {
			tagParts := strings.Split(tag, ",")
			if tagParts[0] == "-" {
				continue // hidden field
			}
			if tagParts[0] != "" {
				continue // field explicitly named
			}
		}
		binding.ToNames = []string{e.translate(binding.Field.Name())}
		binding.FromNames = []string{e.translate(binding.Field.Name())}
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
