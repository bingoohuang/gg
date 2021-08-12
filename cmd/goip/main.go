package main

import (
	"flag"
	"log"
	"net"

	"github.com/bingoohuang/gg/pkg/goip"
)

func main() {
	iface := flag.String("iface", "", "Interface name pattern specified(eg. eth0, eth*)")
	verbose := flag.Bool("verbose", false, "Verbose output for more details")
	v4 := flag.Bool("4", false, "only show ipv4")
	v6 := flag.Bool("6", false, "only show ipv6")
	flag.Parse()

	if !*v4 && !*v6 {
		*v4 = true
	}

	mainIP, ipList := goip.MainIPVerbose(*verbose, *iface)
	log.Printf("Main IP: %s", mainIP)
	log.Printf("IP: %v", ipList)
	log.Printf("Outbound IP: %v", goip.Outbound())

	if *v4 {
		allIPv4, _ := goip.ListAllIPv4(*iface)
		log.Printf("IPv4: %v", allIPv4)
	}

	if *v6 {
		allIPv6, _ := goip.ListAllIPv6(*iface)
		log.Printf("IPv6: %v", allIPv6)
	}

	if *verbose {
		ListIfaces(*v4, *v6, *iface)
		moreInfo()
	}
}

// ListIfaces 根据mode 列出本机所有IP和网卡名称.
func ListIfaces(v4, v6 bool, ifaceName string) {
	list, err := net.Interfaces()
	if err != nil {
		log.Printf("failed to get interfaces, err: %v", err)
		return
	}

	for _, f := range list {
		listIface(f, v4, v6, ifaceName)
	}
}

func listIface(f net.Interface, v4, v6 bool, ifaceName string) {
	matcher := goip.NewIfaceNameMatcher([]string{ifaceName})

	if f.HardwareAddr == nil || f.Flags&net.FlagUp == 0 || f.Flags&net.FlagLoopback == 1 || !matcher.Matches(f.Name) {
		return
	}

	addrs, err := f.Addrs()
	if err != nil {
		log.Printf("\t failed to f.Addrs, × err: %v", err)
		return
	}

	if len(addrs) == 0 {
		return
	}

	got := false
	for _, a := range addrs {
		var netip net.IP
		log.Printf("addr(%T): %s", a, a)
		switch v := a.(type) {
		case *net.IPAddr:
			netip = v.IP
		case *net.IPNet:
			netip = v.IP
		default:
			log.Print("\t\t not .(*net.IPNet) or .(*net.IPNet) ×")
			continue
		}

		if goip.IsIPv4(netip.String()) && !v4 || goip.IsIPv6(netip.String()) && !v6 {
			continue
		}

		if netip.IsLoopback() {
			log.Print("\t\t IsLoopback ×")
			continue
		}

		got = true
	}

	if got {
		log.Printf("\taddrs %+v √", addrs)
	} else {
		log.Printf("\taddrs %+v ×", addrs)
	}
}

func moreInfo() {
	externalIP := goip.External()
	if externalIP == "" {
		return
	}

	log.Printf("公网IP %s", externalIP)
	if eip := net.ParseIP(externalIP); eip != nil {
		result, err := goip.TabaoAPI(externalIP)
		if err != nil {
			log.Printf("TabaoAPI %v", result)
		}
	}

	ipInt := goip.ToDecimal(net.ParseIP(externalIP))
	log.Printf("Convert %s to decimal number(base 10) : %d", externalIP, ipInt)

	ipResult := goip.FromDecimal(ipInt)
	log.Printf("Convert decimal number(base 10) %d to IPv4 address: %v", ipInt, ipResult)

	isBetween := goip.Betweens(net.ParseIP(externalIP), net.ParseIP("0.0.0.0"), net.ParseIP("255.255.255.255"))
	log.Printf("0.0.0.0 isBetween 255.255.255.255 and %s : %v", externalIP, isBetween)

	isPublicIP := goip.IsPublic(net.ParseIP(externalIP))
	log.Printf("%s is public ip: %v ", externalIP, isPublicIP)
}
