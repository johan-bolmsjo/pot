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
func (p *parserBuf) at(n int) byte {
	return p.head[n]
}

// Unconsumed bytes as byte slice.
func (p *parserBuf) bytes() []byte {
	return p.head
}

// UTF8 start character mask.
const utf8Start = 0x080

// Column number at end of consumed data.
func (p *parserBuf) columnNumber() int {
	column := 0
	for i := len(p.tail) - 1; i >= 0; i-- {
		c := rune(p.tail[i])
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
func (p *parserBuf) len() int {
	return len(p.head)
}

// Line number at end of consumed data.
func (p *parserBuf) lineNumber() int {
	return bytes.Count(p.tail, []byte("\n")) + 1
}

// Format an error with location information.
func (p *parserBuf) errorf(format string, a ...interface{}) error {
	loc := fmt.Sprintf("%d:%d: ", p.lineNumber(), p.columnNumber())
	return fmt.Errorf(loc+format, a...)
}

// Split buffer creating a new buffer (consume).
func (p *parserBuf) split(n int) *parserBuf {
	t := *p
	t.head = t.head[:n]
	p.trimBytesLeft(n)
	return &t
}

// Strip the outer block identified by begChar and endChar.
func (p *parserBuf) stripBlock(begChar, endChar byte) {
	p.trimSpace()
	l := p.len()
	if l >= 2 && p.at(0) == begChar && p.at(l-1) == endChar {
		p.trimBytesLeft(1)
		p.trimBytesRight(1)
	}
}

// Trim all bytes (consume).
func (p *parserBuf) trimAll() {
	p.trimBytesLeft(p.len())
}

// Trim bytes from the left (consume).
func (p *parserBuf) trimBytesLeft(n int) {
	p.head = p.head[n:]
	p.tail = p.tail[:len(p.tail)+n]
}

// Trim bytes from the right.
func (p *parserBuf) trimBytesRight(n int) {
	p.head = p.head[:len(p.head)-n]
}

// Trim space from the left (consume).
func (p *parserBuf) trimSpaceLeft() {
	n := bytes.IndexFunc(p.head, func(r rune) bool { return !unicode.IsSpace(r) })
	if n == -1 {
		n = len(p.head)
	}
	p.trimBytesLeft(n)
}

// Trim space from the left (consume) and the right.
func (p *parserBuf) trimSpace() {
	p.trimSpaceLeft()
	p.head = bytes.TrimRightFunc(p.head, unicode.IsSpace)
}
