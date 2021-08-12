package goip

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// External returns  the external IP address.
func External() string {
	ctx, cncl := context.WithTimeout(context.Background(), time.Second*10)
	defer cncl()

	addr := "http://myexternalip.com/raw"
	resp, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil || resp.Body == nil {
		return ""
	}

	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)

	return strings.ReplaceAll(string(content), "\n", "")
}

// ToDecimal converts IP to Decimal
// https://www.ultratools.com/tools/decimalCalc
// nolint gomnd
func ToDecimal(ipnr net.IP) int64 {
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

// FromDecimal converts decimal number(base 10) to IPv4 address.
// https://www.browserling.com/tools/dec-to-ip
func FromDecimal(ipnr int64) net.IP {
	var bs [4]byte

	bs[0] = byte(ipnr & 0xFF)
	bs[1] = byte((ipnr >> 8) & 0xFF)
	bs[2] = byte((ipnr >> 16) & 0xFF)
	bs[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bs[3], bs[2], bs[1], bs[0])
}

// Betweens ...
func Betweens(test, from, to net.IP) bool {
	if from == nil || to == nil || test == nil {
		return false
	}

	from16 := from.To16()
	to16 := to.To16()
	test16 := test.To16()

	if from16 == nil || to16 == nil || test16 == nil {
		return false
	}

	return bytes.Compare(test16, from16) >= 0 && bytes.Compare(test16, to16) <= 0
}

// IsPublic checks a IPv4 address  is a public or not.
func IsPublic(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}

	return false
}
