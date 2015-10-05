package pot

import (
	"bytes"
	"fmt"
	"text/tabwriter"
)

// Helper type used for pretty printing.
// Once an error occurs it silently ignores any write requests.
type printBuf struct {
	buf bytes.Buffer
	tw  *tabwriter.Writer
	es  *errorSink
	tc  byte
}

// Create a new print buffer.
// The error sink can be used to share "abort on first error" behavior with
// other functions. Pass nil as the error sink to let this function create one.
func newPrintBuf(es *errorSink) *printBuf {
	p := new(printBuf)
	p.tw = tabwriter.NewWriter(&p.buf, 4, 0, 1, ' ', 0)
	if es == nil {
		p.es = new(errorSink)
	} else {
		p.es = es
	}
	return p
}

// Returns the buffer as a byte slice or nil if an error has occurred while
// using the buffer.
func (p *printBuf) bytes() []byte {
	if p.es.ok() {
		return p.buf.Bytes()
	}
	return nil
}

// Returns the error stored in the error sink.
func (p *printBuf) error() error {
	return p.es.error()
}

// Returns the error sink used by the buffer.
func (p *printBuf) errorSink() *errorSink {
	return p.es
}

// Flush tabwriter.
// Does nothing if an error has already occurred.
func (p *printBuf) flush() {
	if p.es.ok() {
		p.es.send(p.tw.Flush())
	}
}

// Printf style formatting to buffer.
// Does nothing if an error has already occurred.
// Returns p to allow chaining of write commands.
func (p *printBuf) printf(format string, a ...interface{}) *printBuf {
	if p.es.ok() {
		p.writeTerm()
		_, err := fmt.Fprintf(p.tw, format, a...)
		p.es.send(err)
	}
	return p
}

// Add character c to the buffer when the next printf or write function is
// called. The character is added before printf or write data.
func (p *printBuf) term(c byte) {
	p.tc = c
}

// Write bytes to buffer.
// Does nothing if an error has already occurred.
func (p *printBuf) write(buf []byte) *printBuf {
	if p.es.ok() {
		p.writeTerm()
		_, err := p.tw.Write(buf)
		p.es.send(err)
	}
	return p
}

// Write termination character to buffer if there is one available.
// The termination character is disabled after having been written to the
// buffer. Returns p to allow chaining of write commands.
func (p *printBuf) writeTerm() {
	if p.tc != 0 {
		_, err := p.tw.Write([]byte{p.tc})
		p.es.send(err)
		p.tc = 0
	}
}
