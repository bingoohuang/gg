package snow

import (
	"encoding/binary"
	"net"

	"github.com/bingoohuang/gg/pkg/goip"
)

func defaultIPNodeID() int64 {
	ip, _ := goip.MainIP()
	return ipNodeID(ip)
}

// nolint gomnd
func ipNodeID(ip string) int64 {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return 0
	}

	nodeID := int64(IP2Uint32(parsed) & 0xff)

	return nodeID
}

// IP2Uint32 return a uint32 from an IP.
// nolint gomnd
func IP2Uint32(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}

	return binary.BigEndian.Uint32(ip)
}
