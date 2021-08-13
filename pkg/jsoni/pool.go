package jsoni

import (
	"io"
)

// IteratorPool a thread safe pool of iterators with same configuration
type IteratorPool interface {
	BorrowIterator(data []byte) *Iterator
	ReturnIterator(iter *Iterator)
}

// StreamPool a thread safe pool of streams with same configuration
type StreamPool interface {
	BorrowStream(writer io.Writer) *Stream
	ReturnStream(stream *Stream)
}

func (c *frozenConfig) BorrowStream(writer io.Writer) *Stream {
	stream := c.streamPool.Get().(*Stream)
	stream.Reset(writer)
	return stream
}

func (c *frozenConfig) ReturnStream(stream *Stream) {
	stream.out = nil
	stream.Error = nil
	stream.Attachment = nil
	c.streamPool.Put(stream)
}

func (c *frozenConfig) BorrowIterator(data []byte) *Iterator {
	iter := c.iteratorPool.Get().(*Iterator)
	iter.ResetBytes(data)
	return iter
}

func (c *frozenConfig) ReturnIterator(iter *Iterator) {
	iter.Error = nil
	iter.Attachment = nil
	c.iteratorPool.Put(iter)
}
