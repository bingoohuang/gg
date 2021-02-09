// Package strcase converts strings to various cases. See the conversion table below:
// | Function                          | Result               |
// |-----------------------------------|----------------------|
// | `ToSnake(s)`                      | `any_kind_of_string_v5` |
// | `ToSnakeUpper(s)`                 | `ANY_KIND_OF_STRING_V5` |
// | `ToKebab(s)`                      | `any-kind-of-string-v5` |
// | `ToKebabUpper(s)`                 | `ANY-KIND-OF-STRING5-V5` |
// | `ToDelimited(s, '.')`             | `any.kind.of.string.v5` |
// | `ToDelimitedUpper(s, '.')`        | `ANY.KIND.OF.STRING.V5` |
// | `ToCamel(s)`                      | `AnyKindOfStringV5`    |
// | `ToCamelLower(s)`                 | `anyKindOfStringV5`    |
package strcase
