# help

## cpu.profile

1. start to collect cpu.profile `touch jj.cpu; kill -USR1 69110`
1. after some time, like 10 minutes, repeat above cmd to stop.
1. `go tool pprof -http :9402 cpu.profile` to view

## mem.profile

1. collect mem.profile `touch jj.mem; kill -USR1 6911`
1. `go tool pprof -http :9402 mem.profile` to view 

