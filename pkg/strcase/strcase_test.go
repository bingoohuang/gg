package strcase

import (
	"fmt"
	"testing"
)

func TestDemo(t *testing.T) {
	s := "AnyKind of string v5"
	fmt.Println("ToSnake(s)", ToSnake(s))
	fmt.Println("ToSnakeUpper(s)", ToSnakeUpper(s))
	fmt.Println("ToKebab(s)", ToKebab(s))
	fmt.Println("ToKebabUpper(s)", ToKebabUpper(s))
	fmt.Println("ToDelimited(s, '.')", ToDelimited(s, '.'))
	fmt.Println("ToDelimitedUpper(s, '.')", ToDelimitedUpper(s, '.'))
	fmt.Println("ToCamel(s)", ToCamel(s))
	fmt.Println("ToCamelLower(s)", ToCamelLower(s))
}
