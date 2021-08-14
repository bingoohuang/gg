package jsoni

import "fmt"

// ReadNil reads a json object as nil and
// returns whether it's a nil or not
func (iter *Iterator) ReadNil() (ret bool) {
	c := iter.nextToken()
	if c == 'n' {
		iter.skip3Bytes('u', 'l', 'l') // null
		return true
	}
	iter.unreadByte()
	return false
}

// ReadBool reads a json object as BoolValue
func (iter *Iterator) ReadBool() (ret bool) {
	c := iter.nextToken()
	if c == 't' {
		iter.skip3Bytes('r', 'u', 'e')
		return true
	}
	if c == 'f' {
		iter.skip4Bytes('a', 'l', 's', 'e')
		return false
	}
	iter.ReportError("ReadBool", "expect t or f, but found "+string([]byte{c}))
	return
}

// SkipAndReturnBytes skip next JSON element, and return its content as []byte.
// The []byte can be kept, it is a copy of data.
func (iter *Iterator) SkipAndReturnBytes() []byte {
	iter.startCapture(iter.head)
	iter.Skip()
	return iter.stopCapture()
}

// SkipAndAppendBytes skips next JSON element and appends its content to
// buffer, returning the result.
func (iter *Iterator) SkipAndAppendBytes(buf []byte) []byte {
	iter.startCaptureTo(buf, iter.head)
	iter.Skip()
	return iter.stopCapture()
}

func (iter *Iterator) startCaptureTo(buf []byte, captureStartedAt int) {
	if iter.captured != nil {
		panic("already in capture mode")
	}
	iter.captureStartedAt = captureStartedAt
	iter.captured = buf
}

func (iter *Iterator) startCapture(captureStartedAt int) {
	iter.startCaptureTo(make([]byte, 0, 32), captureStartedAt)
}

func (iter *Iterator) stopCapture() []byte {
	if iter.captured == nil {
		panic("not in capture mode")
	}
	captured := iter.captured
	remaining := iter.buf[iter.captureStartedAt:iter.head]
	iter.captureStartedAt = -1
	iter.captured = nil
	return append(captured, remaining...)
}

// Skip skips a json object and positions to relatively the next json object
func (iter *Iterator) Skip() {
	c := iter.nextToken()
	switch c {
	case '"':
		iter.skipString()
	case 'n':
		iter.skip3Bytes('u', 'l', 'l') // null
	case 't':
		iter.skip3Bytes('r', 'u', 'e') // true
	case 'f':
		iter.skip4Bytes('a', 'l', 's', 'e') // false
	case '0':
		iter.unreadByte()
		iter.ReadFloat32()
	case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		iter.skipNumber()
	case '[':
		iter.skipArray()
	case '{':
		iter.skipObject()
	default:
		iter.ReportError("Skip", fmt.Sprintf("do not know how to skip: %v", c))
		return
	}
}

func (iter *Iterator) skip4Bytes(b1, b2, b3, b4 byte) {
	if iter.readByte() != b1 {
		iter.ReportError("skip4Bytes", fmt.Sprintf("expect %s", string([]byte{b1, b2, b3, b4})))
		return
	}
	if iter.readByte() != b2 {
		iter.ReportError("skip4Bytes", fmt.Sprintf("expect %s", string([]byte{b1, b2, b3, b4})))
		return
	}
	if iter.readByte() != b3 {
		iter.ReportError("skip4Bytes", fmt.Sprintf("expect %s", string([]byte{b1, b2, b3, b4})))
		return
	}
	if iter.readByte() != b4 {
		iter.ReportError("skip4Bytes", fmt.Sprintf("expect %s", string([]byte{b1, b2, b3, b4})))
		return
	}
}

func (iter *Iterator) skip3Bytes(b1, b2, b3 byte) {
	if iter.readByte() != b1 {
		iter.ReportError("skip3Bytes", fmt.Sprintf("expect %s", string([]byte{b1, b2, b3})))
		return
	}
	if iter.readByte() != b2 {
		iter.ReportError("skip3Bytes", fmt.Sprintf("expect %s", string([]byte{b1, b2, b3})))
		return
	}
	if iter.readByte() != b3 {
		iter.ReportError("skip3Bytes", fmt.Sprintf("expect %s", string([]byte{b1, b2, b3})))
		return
	}
}
