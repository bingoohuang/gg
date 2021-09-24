# jsoni

A high-performance 100% compatible drop-in replacement of "encoding/json"

1. keep synchronized with [json-iterator](https://github.com/json-iterator/go) at commit `024077e@2021-09-11T10:17:26+08:00`

## Features enhanced

1. Support converting between int64/uint64 and string values ( used for Javascript lost accuracy for int64/uint64) . eg. `c := jsoni.Config{EscapeHTML: true, Int64AsString: true}.Froze()`
2. Naming strategies for struct can be set on config. eg. `c.RegisterExtension(&extra.NamingStrategyExtension{Translate: strcase.ToCamelLower})`
3. Config.OmitEmptyMapKeys to omit keys whose value is empty.

You can also use thrift like JSON using [thrift-iterator](https://github.com/thrift-iterator/go)

## Benchmark

![benchmark](http://jsoniter.com/benchmarks/go-benchmark.png)

Source code: https://github.com/json-iterator/go-benchmark/blob/master/src/github.com/json-iterator/go-benchmark/benchmark_medium_payload_test.go

Raw Result (easyjson requires static code generation)

|                 | ns/op       | allocation bytes | allocation times |
| --------------- | ----------- | ---------------- | ---------------- |
| std decode      | 35510 ns/op | 1960 B/op        | 99 allocs/op     |
| easyjson decode | 8499 ns/op  | 160 B/op         | 4 allocs/op      |
| jsoniter decode | 5623 ns/op  | 160 B/op         | 3 allocs/op      |
| std encode      | 2213 ns/op  | 712 B/op         | 5 allocs/op      |
| easyjson encode | 883 ns/op   | 576 B/op         | 3 allocs/op      |
| jsoniter encode | 837 ns/op   | 384 B/op         | 4 allocs/op      |

Always benchmark with your own workload.
The result depends heavily on the data input.

# Usage

100% compatibility with standard lib

Replace

```go
import "encoding/json"
json.Marshal(&data)
```

with

```go
import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary
json.Marshal(&data)
```

Replace

```go
import "encoding/json"
json.Unmarshal(input, &data)
```

with

```go
import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary
json.Unmarshal(input, &data)
```

[More documentation](http://jsoniter.com/migrate-from-go-std.html)

# How to get

```
go get github.com/json-iterator/go
```

# Contribution Welcomed !

Contributors

- [thockin](https://github.com/thockin)
- [mattn](https://github.com/mattn)
- [cch123](https://github.com/cch123)
- [Oleg Shaldybin](https://github.com/olegshaldybin)
- [Jason Toffaletti](https://github.com/toffaletti)

Report issue or pull request, or email taowen@gmail.com, or [![Gitter chat](https://badges.gitter.im/gitterHQ/gitter.png)](https://gitter.im/json-iterator/Lobby)
