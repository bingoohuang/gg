package iox

import (
	"bufio"
	"io"
)

type BufioWriteCloser struct {
	closer io.WriteCloser
	*bufio.Writer
}

func NewBufioWriteCloser(w io.WriteCloser) *BufioWriteCloser {
	return &BufioWriteCloser{closer: w, Writer: bufio.NewWriter(w)}
}

func (b *BufioWriteCloser) Close() error {
	_ = b.Writer.Flush()
	return b.closer.Close()
}
