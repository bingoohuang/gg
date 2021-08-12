雪花ID算法Go实现使用手册

1. 安装 `go get github.com/bingoohuang/snow`
1. 使用

   获得类型 | 示例
   ---|---
   字符串| `snow.Next().String()`
   整型   |  `snow.Next().Int64()`
   Bytes | `snow.Next().Bytes()`
   Base2编码 |`snow.Next().Base2()`
   Base32编码 |`snow.Next().Base32()`
   Base36编码 |`snow.Next().Base36()`
   Base58编码| `snow.Next().Base58()`
   Base64编码| `snow.Next().Base64()`

1. 本实现在不指定NodeID的情况下，默认使用主IPv4的最后一部分作为NodeID使用。
