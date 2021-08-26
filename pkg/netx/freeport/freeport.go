package freeport

import (
	"fmt"
	"net"
)

// IsFree tells whether the port is free or not
func IsFree(port int) bool {
	l, err := Listen(port)
	if err != nil {
		return false
	}

	_ = l.Close()
	return true
}

// PortE asks the kernel for a free open port that is ready to use.
func PortE() (int, error) {
	l, e := Listen(0)
	if e != nil {
		return 0, e
	}

	_ = l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// Port Ask the kernel for a free open port that is ready to use.
// this will panic if any error is encountered.
func Port() int {
	port, err := PortE()
	if err != nil {
		panic(err)
	}

	return port
}

// Ports asks the kernel for free open ports that are ready to use.
func Ports(count int) ([]int, error) {
	ports := make([]int, count)

	for i := 0; i < count; i++ {
		p, err := PortE()
		if err != nil {
			return nil, err
		}
		ports[i] = p
	}

	return ports, nil
}

// Listen listens on port.
func Listen(port int) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf(":%d", port))
}

// PortStart finds a free port from starting port
func PortStart(starting int) int {
	p := starting
	for ; !IsFree(p); p++ {
	}
	return p
}
