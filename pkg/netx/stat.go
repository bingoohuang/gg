package netx

import (
	"net"
	"sync/atomic"
)

// StatConn statistics read and write bytes for net connection.
type StatConn struct {
	net.Conn
	Reads, Writes *uint64
}

func NewStatConn(c net.Conn) *StatConn {
	var r, w uint64
	return &StatConn{Conn: c, Reads: &r, Writes: &w}
}

func NewStatConnReadWrite(c net.Conn, reads, writes *uint64) *StatConn {
	return &StatConn{Conn: c, Reads: reads, Writes: writes}
}

// Read reads bytes from net connection.
func (s *StatConn) Read(b []byte) (n int, err error) {
	if n, err = s.Conn.Read(b); err == nil {
		atomic.AddUint64(s.Reads, uint64(n))
	}

	return
}

// Write writes bytes to net.
func (s *StatConn) Write(b []byte) (n int, err error) {
	if n, err = s.Conn.Write(b); err == nil {
		atomic.AddUint64(s.Writes, uint64(n))
	}

	return
}

func (s StatConn) ReadBytes() uint64  { return *s.Reads }
func (s StatConn) WriteBytes() uint64 { return *s.Writes }
