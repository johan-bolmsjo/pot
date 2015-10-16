package pot

import (
	"bytes"
	"unicode"
)

// Helper type that wraps the input text to parse.
// It provides a means to report correct location information in errors.
type parserBuf struct {
	bytes    []byte   // Text to parse
	location Location // Parser location in text input.
}

// Create buffer from byte slice.
func newParserBuf(bytes []byte) *parserBuf {
	return &parserBuf{bytes: bytes}
}

// Format an error with the parser location in text input.
func (buf *parserBuf) errorf(format string, a ...interface{}) error {
	return buf.location.Errorf(format, a...)
}

// Split buffer creating a new buffer.
func (buf *parserBuf) split(n int) *parserBuf {
	t := *buf
	t.bytes = t.bytes[:n]
	buf.trimBytesLeft(n)
	return &t
}

// Strip the outer block identified by begChar and endChar.
func (buf *parserBuf) stripBlock(begChar, endChar byte) {
	buf.trimSpace()
	l := len(buf.bytes)
	if l >= 2 && buf.bytes[0] == begChar && buf.bytes[l-1] == endChar {
		buf.trimBytesLeft(1)
		buf.trimBytesRight(1)
	}
}

// Trim all bytes.
func (buf *parserBuf) trimAll() {
	buf.trimBytesLeft(len(buf.bytes))
}

// Trim bytes from the left.
func (buf *parserBuf) trimBytesLeft(n int) {
	buf.location.updateFromBytes(buf.bytes[:n])
	buf.bytes = buf.bytes[n:]
}

// Trim bytes from the right.
func (buf *parserBuf) trimBytesRight(n int) {
	buf.bytes = buf.bytes[:len(buf.bytes)-n]
}

// Trim space from the left.
func (buf *parserBuf) trimSpaceLeft() {
	n := bytes.IndexFunc(buf.bytes, func(r rune) bool { return !unicode.IsSpace(r) })
	if n == -1 {
		n = len(buf.bytes)
	}
	buf.trimBytesLeft(n)
}

// Trim space from the right.
func (buf *parserBuf) trimSpaceRight() {
	buf.bytes = bytes.TrimRightFunc(buf.bytes, unicode.IsSpace)
}

// Trim space from the left and the right.
func (buf *parserBuf) trimSpace() {
	buf.trimSpaceLeft()
	buf.trimSpaceRight()
}
