package goip

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// IsIPv4 tells a string if in IPv4 format.
func IsIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

// IsIPv6 tells a string if in IPv6 format.
func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

// ListAllIPv4 list all IPv4 addresses.
// ifaceNames are used to specified interface names (filename wild match pattern supported also, like eth*).
func ListAllIPv4(ifaceNames ...string) ([]string, error) {
	ips := make([]string, 0)

	_, err := ListAllIP(func(ip net.IP) (yes bool) {
		s := ip.String()
		if yes = IsIPv4(s); yes {
			ips = append(ips, s)
		}

		return yes
	}, ifaceNames...)

	return ips, err
}

// ListAllIPv6 list all IPv6 addresses.
// ifaceNames are used to specified interface names (filename wild match pattern supported also, like eth*).
func ListAllIPv6(ifaceNames ...string) ([]string, error) {
	ips := make([]string, 0)

	_, err := ListAllIP(func(ip net.IP) (yes bool) {
		s := ip.String()
		if yes = IsIPv6(s); yes {
			ips = append(ips, s)
		}

		return yes
	}, ifaceNames...)

	return ips, err
}

// ListIfaceNames list all net interface names.
func ListIfaceNames() (names []string) {
	list, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for _, i := range list {
		f := i.Flags
		if i.HardwareAddr == nil || f&net.FlagUp == 0 || f&net.FlagLoopback == 1 {
			continue
		}

		names = append(names, i.Name)
	}

	return names
}

// ListAllIP list all IP addresses.
func ListAllIP(predicate func(net.IP) bool, ifaceNames ...string) ([]net.IP, error) {
	list, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces, err: %w", err)
	}

	ips := make([]net.IP, 0)
	matcher := NewIfaceNameMatcher(ifaceNames)

	for _, i := range list {
		f := i.Flags
		if i.HardwareAddr == nil ||
			f&net.FlagUp != net.FlagUp ||
			f&net.FlagLoopback == net.FlagLoopback ||
			!matcher.Matches(i.Name) {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		ips = collectAddresses(predicate, addrs, ips)
	}

	return ips, nil
}

func collectAddresses(predicate func(net.IP) bool, addrs []net.Addr, ips []net.IP) []net.IP {
	for _, a := range addrs {
		var ip net.IP
		switch v := a.(type) {
		case *net.IPAddr:
			ip = v.IP
		case *net.IPNet:
			ip = v.IP
		default:
			continue
		}

		if !ContainsIS(ips, ip) && predicate(ip) {
			ips = append(ips, ip)
		}
	}

	return ips
}

func ContainsIS(ips []net.IP, ip net.IP) bool {
	for _, j := range ips {
		if j.Equal(ip) {
			return true
		}
	}

	return false
}

// Outbound  gets preferred outbound ip of this machine.
func Outbound() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}

	defer conn.Close()

	s := conn.LocalAddr().String()
	return s[:strings.LastIndex(s, ":")]
}

// MainIP tries to get the main IP address and the IP addresses.
func MainIP(ifaceName ...string) (string, []string) {
	return MainIPVerbose(false, ifaceName...)
}

// MainIPVerbose tries to get the main IP address and the IP addresses.
func MainIPVerbose(verbose bool, ifaceName ...string) (string, []string) {
	ips, _ := ListAllIPv4(ifaceName...)
	if len(ips) == 1 {
		return ips[0], ips
	}

	if s := findMainIPByIfconfig(verbose, ifaceName); s != "" {
		return s, ips
	}

	if out := Outbound(); out != "" && contains(ips, out) {
		return out, ips
	}

	if len(ips) > 0 {
		return ips[0], ips
	}

	return "", nil
}

func findMainIPByIfconfig(verbose bool, ifaceName []string) string {
	names := ListIfaceNames()
	if verbose {
		log.Printf("iface names: %s", names)
	}

	var matchedNames []string
	matcher := NewIfaceNameMatcher(ifaceName)
	for _, n := range names {
		if matcher.Matches(n) {
			matchedNames = append(matchedNames, n)
		}
	}

	if verbose && len(matchedNames) < len(names) {
		log.Printf("matchedNames: %s", matchedNames)
	}

	if len(matchedNames) == 0 {
		return ""
	}

	name := matchedNames[0]
	for _, n := range matchedNames {
		// for en0 on mac or eth0 on linux
		if strings.HasPrefix(n, "e") && strings.HasSuffix(n, "0") {
			name = n
			break
		}
	}

	/*
		[root@tencent-beta17 ~]# ifconfig eth0
		eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
		        inet 192.168.108.7  netmask 255.255.255.0  broadcast 192.168.108.255
		        ether 52:54:00:ef:16:bd  txqueuelen 1000  (Ethernet)
		        RX packets 1838617728  bytes 885519190162 (824.7 GiB)
		        RX errors 0  dropped 0  overruns 0  frame 0
		        TX packets 1665532349  bytes 808544539610 (753.0 GiB)
		        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
	*/
	re := regexp.MustCompile(`inet\s+([\w.]+?)\s+`)
	if verbose {
		log.Printf("exec comd: ifconfig %s", name)
	}
	c := exec.Command("ifconfig", name)
	if co, err := c.Output(); err == nil {
		if verbose {
			log.Printf("output: %s", co)
		}
		sub := re.FindStringSubmatch(string(co))
		if len(sub) > 1 {
			if verbose {
				log.Printf("found: %s", sub[1])
			}
			return sub[1]
		}
	} else if verbose {
		log.Printf("error: %v", err)
	}

	return ""
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}

	return false
}

// MakeSliceMap makes a map[string]bool from the string slice.
func MakeSliceMap(ss []string) map[string]bool {
	m := make(map[string]bool)

	for _, s := range ss {
		if s != "" {
			m[s] = true
		}
	}

	return m
}

type IfaceNameMatcher struct {
	ifacePatterns map[string]bool
}

func NewIfaceNameMatcher(ss []string) IfaceNameMatcher {
	return IfaceNameMatcher{ifacePatterns: MakeSliceMap(ss)}
}

func (i IfaceNameMatcher) Matches(name string) bool {
	if len(i.ifacePatterns) == 0 {
		return true
	}

	if _, ok := i.ifacePatterns[name]; ok {
		return true
	}

	for k := range i.ifacePatterns {
		if ok, _ := filepath.Match(k, name); ok {
			return true
		}
	}

	for k := range i.ifacePatterns {
		if strings.Contains(k, name) {
			return true
		}
	}

	return false
}
