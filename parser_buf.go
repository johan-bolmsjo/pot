package pot

import (
	"bytes"
	"fmt"
	"unicode"
)

// Helper type that wraps the input text to parse.
// It provides a means to report correct location information in errors
// and to extract the original input if needed.
//
type parserBuf struct {
	head []byte // Data to parse
	tail []byte // Consumed data
}

// Create buffer from byte slice.
func newParserBuf(text []byte) *parserBuf {
	return &parserBuf{text, text[:0]}
}

// Byte at index.
func (buf *parserBuf) at(n int) byte {
	return buf.head[n]
}

// Unconsumed bytes as byte slice.
func (buf *parserBuf) bytes() []byte {
	return buf.head
}

// UTF8 start character mask.
const utf8Start = 0x080

// Column number at end of consumed data.
func (buf *parserBuf) columnNumber() int {
	column := 0
	for i := len(buf.tail) - 1; i >= 0; i-- {
		c := rune(buf.tail[i])
		if c == '\r' || c == '\n' {
			break
		}
		if c&utf8Start == 0 {
			column++
		}
	}
	return column
}

// Number of unconsumed bytes.
func (buf *parserBuf) len() int {
	return len(buf.head)
}

// Line number at end of consumed data.
func (buf *parserBuf) lineNumber() int {
	return bytes.Count(buf.tail, []byte("\n")) + 1
}

// Format an error with location information.
func (buf *parserBuf) errorf(format string, a ...interface{}) error {
	loc := fmt.Sprintf("%d:%d: ", buf.lineNumber(), buf.columnNumber())
	return fmt.Errorf(loc+format, a...)
}

// Split buffer creating a new buffer (consume).
func (buf *parserBuf) split(n int) *parserBuf {
	t := *buf
	t.head = t.head[:n]
	buf.trimBytesLeft(n)
	return &t
}

// Strip the outer block identified by begChar and endChar.
func (buf *parserBuf) stripBlock(begChar, endChar byte) {
	buf.trimSpace()
	l := buf.len()
	if l >= 2 && buf.at(0) == begChar && buf.at(l-1) == endChar {
		buf.trimBytesLeft(1)
		buf.trimBytesRight(1)
	}
}

// Trim all bytes (consume).
func (buf *parserBuf) trimAll() {
	buf.trimBytesLeft(buf.len())
}

// Trim bytes from the left (consume).
func (buf *parserBuf) trimBytesLeft(n int) {
	buf.head = buf.head[n:]
	buf.tail = buf.tail[:len(buf.tail)+n]
}

// Trim bytes from the right.
func (buf *parserBuf) trimBytesRight(n int) {
	buf.head = buf.head[:len(buf.head)-n]
}

// Trim space from the left (consume).
func (buf *parserBuf) trimSpaceLeft() {
	n := bytes.IndexFunc(buf.head, func(r rune) bool { return !unicode.IsSpace(r) })
	if n == -1 {
		n = len(buf.head)
	}
	buf.trimBytesLeft(n)
}

// Trim space from the left (consume) and the right.
func (buf *parserBuf) trimSpace() {
	buf.trimSpaceLeft()
	buf.head = bytes.TrimRightFunc(buf.head, unicode.IsSpace)
}
