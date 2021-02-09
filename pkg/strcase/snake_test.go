package strcase

import (
	"testing"
)

type SnakeTest struct {
	input  string
	output string
}

// nolint gochecknoglobals
var tests = []SnakeTest{
	{"woof0_woof1", "woof0_woof1"},
	{"_woof0_woof1_2", "_woof0_woof1_2"},
	{"woof0_WOOF1_2", "woof0_woof1_2"},

	{"a", "a"},
	{"snake", "snake"},
	{"A", "a"},
	{"ID", "id"},
	{"MOTD", "motd"},
	{"Snake", "snake"},
	{"SnakeTest", "snake_test"},
	{"APIResponse", "api_response"},
	{"SnakeID", "snake_id"},
	{"Snake_Id", "snake_id"},
	{"Snake_ID", "snake_id"},
	{"SnakeIDGoogle", "snake_id_google"},
	{"LinuxMOTD", "linux_motd"},
	{"OMGWTFBBQ", "omgwtfbbq"},
	{"omg_wtf_bbq", "omg_wtf_bbq"},
	{"woof_woof", "woof_woof"},
	{"_woof_woof", "_woof_woof"},
	{"woof_woof_", "woof_woof_"},
	{"WOOF", "woof"},
	{"Woof", "woof"},
	{"woof", "woof"},

	{"WOOF0", "woof0"},
	{"Woof1", "woof1"},
	{"woof2", "woof2"},
	{"woofWoof", "woof_woof"},
	{"woofWOOF", "woof_woof"},
	{"woof_WOOF", "woof_woof"},
	{"Woof_WOOF", "woof_woof"},
	{"WOOFWoofWoofWOOFWoofWoof", "woof_woof_woof_woof_woof_woof"},
	{"WOOF_Woof_woof_WOOF_Woof_woof", "woof_woof_woof_woof_woof_woof"},
	{"Woof_W", "woof_w"},
	{"Woof_w", "woof_w"},
	{"WoofW", "woof_w"},
	{"Woof_W_", "woof_w_"},
	{"Woof_w_", "woof_w_"},
	{"WoofW_", "woof_w_"},
	{"WOOF_", "woof_"},
	{"W_Woof", "w_woof"},
	{"w_Woof", "w_woof"},
	{"WWoof", "w_woof"},
	{"_W_Woof", "_w_woof"},
	{"_w_Woof", "_w_woof"},
	{"_WWoof", "_w_woof"},
	{"_WOOF", "_woof"},
	{"_woof", "_woof"},
	{"Load5", "load5"},
	{"V1", "v1"},
}

func TestSnakeCase(t *testing.T) {
	for _, test := range tests {
		if ToSnake(test.input) != test.output {
			t.Errorf("SnakeCase(%q) -> %q, want %q", test.input, ToSnake(test.input), test.output)
		}
	}
}

func BenchmarkSnakeCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			ToSnake(test.input)
		}
	}
}

func TestToSnake(t *testing.T) {
	cases := [][]string{
		{"v1", "v1"},
		{"testCase", "test_case"},
		{"TestCase", "test_case"},
		{"Test Case", "test_case"},
		{" Test Case", "test_case"},
		{"Test Case ", "test_case"},
		{" Test Case ", "test_case"},
		{"test", "test"},
		{"test_case", "test_case"},
		{"Test", "test"},
		{"", ""},
		{"ManyManyWords", "many_many_words"},
		{"manyManyWords", "many_many_words"},
		{"AnyKind of_string", "any_kind_of_string"},
		{"numbers2and55with000", "numbers2_and55_with000"},
		{"JSONData", "json_data"},
		{"userID", "user_id"},
		{"AAAbbb", "aa_abbb"},
	}
	for _, i := range cases {
		in := i[0]
		out := i[1]
		result := ToSnake(in)

		if result != out {
			t.Error("'" + in + "'('" + result + "' != '" + out + "')")
		}
	}
}

func TestToDelimited(t *testing.T) {
	cases := [][]string{
		{"testCase", "test@case"},
		{"TestCase", "test@case"},
		{"Test Case", "test@case"},
		{" Test Case", "test@case"},
		{"Test Case ", "test@case"},
		{" Test Case ", "test@case"},
		{"test", "test"},
		{"test_case", "test@case"},
		{"Test", "test"},
		{"", ""},
		{"ManyManyWords", "many@many@words"},
		{"manyManyWords", "many@many@words"},
		{"AnyKind of_string", "any@kind@of@string"},
		{"numbers2and55with000", "numbers2@and55@with000"},
		{"JSONData", "json@data"},
		{"userID", "user@id"},
		{"AAAbbb", "aa@abbb"},
		{"test-case", "test@case"},
	}
	for _, i := range cases {
		in := i[0]
		out := i[1]

		result := ToDelimited(in, '@')
		if result != out {
			t.Error("'" + in + "' ('" + result + "' != '" + out + "')")
		}
	}
}

func TestToSnakeUpper(t *testing.T) {
	cases := [][]string{
		{"testCase", "TEST_CASE"},
	}
	for _, i := range cases {
		in := i[0]
		out := i[1]

		result := ToSnakeUpper(in)
		if result != out {
			t.Error("'" + result + "' != '" + out + "'")
		}
	}
}

func TestToKebab(t *testing.T) {
	cases := [][]string{
		{"v1", "v1"},
		{"testCase", "test-case"},
	}
	for _, i := range cases {
		in := i[0]
		out := i[1]

		result := ToKebab(in)
		if result != out {
			t.Error("'" + result + "' != '" + out + "'")
		}
	}
}

func TestToKebabUpper(t *testing.T) {
	cases := [][]string{
		{"testCase", "TEST-CASE"},
	}
	for _, i := range cases {
		in := i[0]
		out := i[1]
		result := ToKebabUpper(in)

		if result != out {
			t.Error("'" + result + "' != '" + out + "'")
		}
	}
}

func TestToDelimitedUpper(t *testing.T) {
	cases := [][]string{
		{"testCase", "TEST.CASE"},
	}
	for _, i := range cases {
		in := i[0]
		out := i[1]
		result := ToDelimitedUpper(in, '.')

		if result != out {
			t.Error("'" + result + "' != '" + out + "'")
		}
	}
}
