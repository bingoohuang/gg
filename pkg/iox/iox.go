package iox

import (
	"bufio"
	"io"
	"log"
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

func ReadString(r io.Reader) string {
	return string(ReadBytes(r))
}

func ReadBytes(r io.Reader) []byte {
	data, err := io.ReadAll(r)
	if err != nil {
		log.Printf("read bytes failed: %v", err)
	}

	return data
}

// DiscardClose discards the reader and then close it.
func DiscardClose(c io.ReadCloser) {
	if c != nil {
		_, _ = io.Copy(io.Discard, c)
		Close(c)
	}
}
