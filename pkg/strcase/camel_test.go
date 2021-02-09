package strcase

import (
	"testing"
)

func TestToCamel(t *testing.T) {
	cases := [][]string{
		{"v1", "V1"},
		{"test_case", "TestCase"},
		{"test_CASE", "TestCase"},
		{"test", "Test"},
		{"TestCase", "TestCase"},
		{"mySQL", "MySql"},
		{" test  case ", "TestCase"},
		{"", ""},
		{"many_many_words", "ManyManyWords"},
		{"AnyKind of_string", "AnyKindOfString"},
		{"odd-fix", "OddFix"},
		{"numbers2And55with000", "Numbers2And55With000"},
	}
	for _, i := range cases {
		in := i[0]
		out := i[1]
		result := ToCamel(in)

		if result != out {
			t.Error("'" + result + "' != '" + out + "'")
		}
	}
}

func TestToCamelLower(t *testing.T) {
	cases := [][]string{
		{"ID", "id"},
		{"SQL", "sql"},
		{"MySQL", "mySql"},
		{"SQLMap", "sqlMap"},
		{"foo-bar", "fooBar"},
		{"foo_bar", "fooBar"},
		{"TestCase", "testCase"},
		{"", ""},
		{"AnyKind of_string", "anyKindOfString"},
	}

	for _, i := range cases {
		in := i[0]
		out := i[1]
		result := ToCamelLower(in)

		if result != out {
			t.Error("'" + result + "' != '" + out + "'")
		}
	}
}
