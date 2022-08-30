# strcase

forked from https://github.com/iancoleman/strcase

strcase is a go package for converting string to various cases (e.g. [snake case](https://en.wikipedia.org/wiki/Snake_case) or [camel case](https://en.wikipedia.org/wiki/CamelCase)) to see the full conversion table below.

## Example


| s                      | Function                   | Result                   |
|------------------------|----------------------------|--------------------------|
| `AnyKind of string v5` | `ToSnake(s)`               | `any_kind_of_string_v5`  |
| `AnyKind of string v5` | `ToSnakeUpper(s)`          | `ANY_KIND_OF_STRING_V5`  |
| `AnyKind of string v5` | `ToKebab(s)`               | `any-kind-of-string-v5`  |
| `AnyKind of string v5` | `ToKebabUpper(s)`          | `ANY-KIND-OF-STRING5-V5` |
| `AnyKind of string v5` | `ToDelimited(s, '.')`      | `any.kind.of.string.v5`  |
| `AnyKind of string v5` | `ToDelimitedUpper(s, '.')` | `ANY.KIND.OF.STRING.V5`  |
| `AnyKind of string v5` | `ToCamel(s)`               | `AnyKindOfStringV5`      |
| `mySQL`                | `ToCamel(s)`               | `MySql`                  |
| `AnyKind of string v5` | `ToCamelLower(s)`          | `anyKindOfStringV5`      |
| `ID`                   | `ToCamelLower(s)`          | `id`                     |
| `SQLMap`               | `ToCamelLower(s)`          | `sqlMap`                 |
| `TestCase`             | `ToCamelLower(s)`          | `fooBar`                 |
| `foo-bar`              | `ToCamelLower(s)`          | `fooBar`                 |
| `foo_bar`              | `ToCamelLower(s)`          | `fooBar`                 |


case conversion types:

- Camel Case (e.g. CamelCase)
- Lower Camel Case (e.g. lowerCamelCase)
- Snake Case (e.g. snake_case)
- Screaming Snake Case (e.g. SCREAMING_SNAKE_CASE)
- Kebab Case (e.g. kebab-case)
- Screaming Kebab Case(e.g. SCREAMING-KEBAB-CASE)
- Dot Notation Case (e.g. dot.notation.case)
- Screaming Dot Notation Case (e.g. DOT.NOTATION.CASE)
- Title Case (e.g. Title Case)
- Other delimiters

## resources

1. [caps a case conversion library for Go](https://github.com/chanced/caps)
