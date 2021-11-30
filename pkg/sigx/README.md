# help

## 代码嵌入

```go
package main

import (
	"github.com/bingoohuang/gg/pkg/sigx"
)

func init() {
	sigx.RegisterSignalProfile()
}
```

## cpu.profile

1. start to collect cpu.profile `touch jj.cpu; kill -USR1 69110`
1. after some time, like 10 minutes, repeat above cmd to stop.
1. `go tool pprof -http :9402 cpu.profile` to view

## mem.profile

1. collect mem.profile `touch jj.mem; kill -USR1 6911`
1. `go tool pprof -http :9402 mem.profile` to view

## jj.profile

1. collect specified `echo "cpu,heap,allocs,mutex,block,trace,threadcreate,goroutine,d:5m,rate:4096" > jj.profile`
2. `go tool pprof -http :9402 xx.pprof` or `go tool trace trace.out` to view
