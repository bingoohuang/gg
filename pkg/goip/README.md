# ip

[![Travis CI](https://img.shields.io/travis/bingoohuang/ip/master.svg?style=flat-square)](https://travis-ci.com/bingoohuang/ip)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/bingoohuang/ip/blob/master/LICENSE.md)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/bingoohuang/ip)
[![Coverage Status](http://codecov.io/github/bingoohuang/ip/coverage.svg?branch=master)](http://codecov.io/github/bingoohuang/ip?branch=master)
[![goreport](https://www.goreportcard.com/badge/github.com/bingoohuang/ip)](https://www.goreportcard.com/report/github.com/bingoohuang/ip)

show host IP addresses

## API

```go
import "github.com/bingoohuang/ip"

// ListAllIPv4 list all IPv4 addresses.
// The input argument ifaceNames are used to specified interface names
// (filename wild match pattern supported also, like eth*)
allIPv4s, _ := ip.ListAllIPv4()
allEth0IPv4s, _ := ip.ListAllIPv4("eth0")
allEth0En0IPv4s, _ := ip.ListAllIPv4("eth0", "en0")
allEnIPv4s, _ := ip.ListAllIPv4("en*")

// GetOutboundIP  gets preferred outbound ip of this machine.
outboundIP := ip.Outbound()

// MainIP tries to get the main IP address and the IP addresses.
mainIP, ipList := ip.MainIP()
```

## Usages

on mac:

```bash
$ ip
INFO[0000] MainIP: 192.168.162.41, IP List: [192.168.162.41]
INFO[0000] OutboundIP: 192.168.162.41
INFO[0015] 公网IP 60.247.93.190
INFO[0015] TabaoAPI &{Code:0 Data:{Country:中国 CountryID:CN Area: AreaID: Region:北京 RegionID:110000 City:北京 CityID:110100 Isp:电信}}
INFO[0015] Convert 60.247.93.190 to decimal number(base 10) : 1022844350
INFO[0015] Convert decimal number(base 10) 1022844350 to IPv4 address: 60.247.93.190
INFO[0015] 0.0.0.0 isBetween 255.255.255.255 and 60.247.93.190 : true
INFO[0015] 60.247.93.190 is public ip: true
INFO[0015] iface {Index:1 MTU:16384 Name:lo0 HardwareAddr: Flags:up|loopback|multicast}
INFO[0015] iface {Index:2 MTU:1280 Name:gif0 HardwareAddr: Flags:pointtopoint|multicast}
INFO[0015] iface {Index:3 MTU:1280 Name:stf0 HardwareAddr: Flags:0}
INFO[0015] iface {Index:4 MTU:0 Name:XHC0 HardwareAddr: Flags:0}
INFO[0015] iface {Index:5 MTU:0 Name:XHC1 HardwareAddr: Flags:0}
INFO[0015] iface {Index:6 MTU:0 Name:XHC20 HardwareAddr: Flags:0}
INFO[0015] iface {Index:7 MTU:0 Name:VHC128 HardwareAddr: Flags:0}
INFO[0015] iface {Index:8 MTU:1500 Name:en5 HardwareAddr:ac:de:48:00:11:22 Flags:up|broadcast|multicast}
INFO[0015]      addrs [fe80::aede:48ff:fe00:1122/64]
INFO[0015]              √ Got fe80::aede:48ff:fe00:1122
INFO[0015] iface {Index:9 MTU:1500 Name:ap1 HardwareAddr:f2:18:98:a5:12:27 Flags:broadcast|multicast}
INFO[0015] iface {Index:10 MTU:1500 Name:en0 HardwareAddr:f0:18:98:a5:12:27 Flags:up|broadcast|multicast}
INFO[0015]      addrs [fe80::413:5cf0:7a62:facf/64 192.168.162.41/24]
INFO[0015]              √ Got fe80::413:5cf0:7a62:facf
INFO[0015]              √ Got 192.168.162.41
INFO[0015] iface {Index:11 MTU:1500 Name:en1 HardwareAddr:82:68:2b:61:34:01 Flags:up|broadcast|multicast}
INFO[0015] iface {Index:12 MTU:1500 Name:en2 HardwareAddr:82:68:2b:61:34:00 Flags:up|broadcast|multicast}
INFO[0015] iface {Index:13 MTU:1500 Name:en3 HardwareAddr:82:68:2b:61:34:05 Flags:up|broadcast|multicast}
INFO[0015] iface {Index:14 MTU:1500 Name:en4 HardwareAddr:82:68:2b:61:34:04 Flags:up|broadcast|multicast}
INFO[0015] iface {Index:15 MTU:1500 Name:bridge0 HardwareAddr:82:68:2b:61:34:01 Flags:up|broadcast|multicast}
INFO[0015] iface {Index:16 MTU:2304 Name:p2p0 HardwareAddr:02:18:98:a5:12:27 Flags:up|broadcast|multicast}
INFO[0015] iface {Index:17 MTU:1484 Name:awdl0 HardwareAddr:02:62:29:9b:38:bf Flags:up|broadcast|multicast}
INFO[0015]      addrs [fe80::62:29ff:fe9b:38bf/64]
INFO[0015]              √ Got fe80::62:29ff:fe9b:38bf
INFO[0015] iface {Index:18 MTU:1500 Name:llw0 HardwareAddr:02:62:29:9b:38:bf Flags:up|broadcast|multicast}
INFO[0015]      addrs [fe80::62:29ff:fe9b:38bf/64]
INFO[0015]              √ Got fe80::62:29ff:fe9b:38bf
INFO[0015] iface {Index:19 MTU:1380 Name:utun0 HardwareAddr: Flags:up|pointtopoint|multicast}
INFO[0015] iface {Index:20 MTU:2000 Name:utun1 HardwareAddr: Flags:up|pointtopoint|multicast}
```

on linux:

1. `make docker`
1. `bssh scp -H q7 ~/dockergo/bin/ip-v1.0.0-amd64-glibc2.28.gz r:bingoohuang`

```bash
$ ./ip-v1.0.0-amd64-glibc2.28
INFO[0000] MainIP: 192.168.1.7, IP List: [192.168.1.7 192.168.1.17 172.18.0.1 172.17.0.1 172.19.0.1 10.42.2.0]
INFO[0000] OutboundIP: 192.168.1.7
INFO[0000] 公网IP 123.206.1.162
INFO[0000] TabaoAPI &{Code:0 Data:{Country:中国 CountryID:CN Area: AreaID: Region:上海 RegionID:310000 City:上海 CityID:310100 Isp:电信}}
INFO[0000] Convert 123.206.1.162 to decimal number(base 10) : 2077145506
INFO[0000] Convert decimal number(base 10) 2077145506 to IPv4 address: 123.206.1.162
INFO[0000] 0.0.0.0 isBetween 255.255.255.255 and 123.206.1.162 : true
INFO[0000] 123.206.1.162 is public ip: true
INFO[0000] iface {Index:1 MTU:65536 Name:lo HardwareAddr: Flags:up|loopback}
INFO[0000] iface {Index:2 MTU:1500 Name:eth0 HardwareAddr:52:54:00:ef:16:bd Flags:up|broadcast|multicast}
INFO[0000]      addrs [192.168.1.7/24 192.168.1.17/32]
INFO[0000]              √ Got 192.168.1.7
INFO[0000]              √ Got 192.168.1.17
INFO[0000] iface {Index:3 MTU:1500 Name:br-8983f91a1c88 HardwareAddr:02:42:49:a5:88:9f Flags:up|broadcast|multicast}
INFO[0000]      addrs [172.18.0.1/16]
INFO[0000]              √ Got 172.18.0.1
INFO[0000] iface {Index:4 MTU:1500 Name:docker0 HardwareAddr:02:42:b3:f8:ea:89 Flags:up|broadcast|multicast}
INFO[0000]      addrs [172.17.0.1/16]
INFO[0000]              √ Got 172.17.0.1
INFO[0000] iface {Index:7 MTU:1500 Name:br-d4979d31f397 HardwareAddr:02:42:65:78:5e:15 Flags:up|broadcast|multicast}
INFO[0000]      addrs [172.19.0.1/16]
INFO[0000]              √ Got 172.19.0.1
INFO[0000] iface {Index:284 MTU:1450 Name:flannel.1 HardwareAddr:ce:1d:c7:1e:20:36 Flags:up|broadcast|multicast}
INFO[0000]      addrs [10.42.2.0/32]
INFO[0000]              √ Got 10.42.2.0
INFO[0000] iface {Index:287 MTU:1500 Name:calif3053f9c1b9 HardwareAddr:ee:ee:ee:ee:ee:ee Flags:up|broadcast|multicast}
INFO[0000] iface {Index:288 MTU:1500 Name:calie04efa792c4 HardwareAddr:ee:ee:ee:ee:ee:ee Flags:up|broadcast|multicast}
INFO[0000] iface {Index:33 MTU:1500 Name:vethb98680d HardwareAddr:ce:8b:3c:1c:c1:9a Flags:up|broadcast|multicast}
INFO[0000] iface {Index:37 MTU:1500 Name:veth69f8b3d HardwareAddr:3e:1d:13:80:f0:a7 Flags:up|broadcast|multicast}
INFO[0000] iface {Index:7726 MTU:1500 Name:calibd36e9f9137 HardwareAddr:ee:ee:ee:ee:ee:ee Flags:up|broadcast|multicast}
INFO[0000] iface {Index:7752 MTU:1500 Name:cali76dffd9c471 HardwareAddr:ee:ee:ee:ee:ee:ee Flags:up|broadcast|multicast}
INFO[0000] iface {Index:7674 MTU:1500 Name:cali94579f886d8 HardwareAddr:ee:ee:ee:ee:ee:ee Flags:up|broadcast|multicast}
INFO[0000] iface {Index:7679 MTU:1500 Name:cali59741a8250f HardwareAddr:ee:ee:ee:ee:ee:ee Flags:up|broadcast|multicast}
```

## List available network interfaces:

`tcpdump -D`

```bash
$ tcpdump -D
1.en0 [Up, Running]
2.p2p0 [Up, Running]
3.awdl0 [Up, Running]
4.llw0 [Up, Running]
5.utun0 [Up, Running]
6.utun1 [Up, Running]
7.lo0 [Up, Running, Loopback]
8.bridge0 [Up, Running]
9.en1 [Up, Running]
10.en2 [Up, Running]
11.en3 [Up, Running]
12.en4 [Up, Running]
13.gif0 [none]
14.stf0 [none]
15.XHC0 [none]
16.XHC1 [none]
17.ap1 [none]
18.XHC20 [none]
19.VHC128 [none]
```

```bash
# tcpdump -D
1.eth0
2.docker0
3.nflog (Linux netfilter log (NFLOG) interface)
4.nfqueue (Linux netfilter queue (NFQUEUE) interface)
5.flannel.1
6.usbmon1 (USB bus number 1)
7.calie04efa792c4
8.calibd36e9f9137
9.veth69f8b3d
10.cali76dffd9c471
11.calif3053f9c1b9
12.br-d4979d31f397
13.br-8983f91a1c88
14.cali59741a8250f
15.cali94579f886d8
16.vethb98680d
17.any (Pseudo-device that captures on all interfaces)
18.lo [Loopback]
```

```bash
$ hostname -I
192.168.1.7 192.168.1.17 172.18.0.1 172.17.0.1 172.19.0.1 10.42.2.0
```

```shell
[root@tencent-beta17 ~]# ifconfig eth0
eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 192.168.108.7  netmask 255.255.255.0  broadcast 192.168.108.255
        ether 52:54:00:ef:16:bd  txqueuelen 1000  (Ethernet)
        RX packets 1838617728  bytes 885519190162 (824.7 GiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 1665532349  bytes 808544539610 (753.0 GiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
[root@tencent-beta17 ~]# ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP qlen 1000
    link/ether 52:54:00:ef:16:bd brd ff:ff:ff:ff:ff:ff
    inet 192.168.108.7/24 brd 192.168.108.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet 192.168.108.17/32 brd 192.168.108.255 scope global eth0
       valid_lft forever preferred_lft forever
3: docker0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN
    link/ether 02:42:f7:47:e6:02 brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.1/16 brd 172.17.255.255 scope global docker0
       valid_lft forever preferred_lft forever
4: br-d4979d31f397: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:17:47:99:97 brd ff:ff:ff:ff:ff:ff
    inet 172.19.0.1/16 brd 172.19.255.255 scope global br-d4979d31f397
       valid_lft forever preferred_lft forever
14: flannel.1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UNKNOWN
    link/ether 2e:09:7c:c5:b0:83 brd ff:ff:ff:ff:ff:ff
    inet 10.42.2.0/32 scope global flannel.1
       valid_lft forever preferred_lft forever
849: veth120ec58@if848: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master br-d4979d31f397 state UP
    link/ether 42:33:d1:8a:ed:55 brd ff:ff:ff:ff:ff:ff link-netnsid 1
851: veth62f525c@if850: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master br-d4979d31f397 state UP
    link/ether 62:eb:7f:58:ac:f3 brd ff:ff:ff:ff:ff:ff link-netnsid 0
852: cali84fc7ac85e8@if3: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UP
    link/ether ee:ee:ee:ee:ee:ee brd ff:ff:ff:ff:ff:ff link-netnsid 3
```

mac:

```sh
➜  ip git:(master) ✗ ifconfig en0
en0: flags=8863<UP,BROADCAST,SMART,RUNNING,SIMPLEX,MULTICAST> mtu 1500
	options=400<CHANNEL_IO>
	ether f0:18:98:a5:12:27
	inet6 fe80::40d:d0ee:8501:f4a7%en0 prefixlen 64 secured scopeid 0xa
	inet 192.168.162.21 netmask 0xffffff00 broadcast 192.168.162.255
	nd6 options=201<PERFORMNUD,DAD>
	media: autoselect
	status: active
```

## thanks to the Giant Shoulders

1. [A minimalist HTTP headers inspector](http://gethttp.info/)
