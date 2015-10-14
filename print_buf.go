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
	pbuf := new(printBuf)
	pbuf.tw = tabwriter.NewWriter(&pbuf.buf, 4, 0, 1, ' ', 0)
	if es == nil {
		pbuf.es = new(errorSink)
	} else {
		pbuf.es = es
	}
	return pbuf
}

// Returns the buffer as a byte slice or nil if an error has occurred while
// using the buffer.
func (pbuf *printBuf) bytes() []byte {
	if pbuf.es.ok() {
		return pbuf.buf.Bytes()
	}
	return nil
}

// Returns the error stored in the error sink.
func (pbuf *printBuf) error() error {
	return pbuf.es.error()
}

// Returns the error sink used by the buffer.
func (pbuf *printBuf) errorSink() *errorSink {
	return pbuf.es
}

// Flush tabwriter.
// Does nothing if an error has already occurred.
func (pbuf *printBuf) flush() {
	if pbuf.es.ok() {
		pbuf.es.send(pbuf.tw.Flush())
	}
}

// Printf style formatting to buffer.
// Does nothing if an error has already occurred.
// Returns p to allow chaining of write commands.
func (pbuf *printBuf) printf(format string, a ...interface{}) *printBuf {
	if pbuf.es.ok() {
		pbuf.writeTerm()
		_, err := fmt.Fprintf(pbuf.tw, format, a...)
		pbuf.es.send(err)
	}
	return pbuf
}

// Add character c to the buffer when the next printf or write function is
// called. The character is added before printf or write data.
func (pbuf *printBuf) term(c byte) {
	pbuf.tc = c
}

// Write bytes to buffer.
// Does nothing if an error has already occurred.
func (pbuf *printBuf) write(buf []byte) *printBuf {
	if pbuf.es.ok() {
		pbuf.writeTerm()
		_, err := pbuf.tw.Write(buf)
		pbuf.es.send(err)
	}
	return pbuf
}

// Write termination character to buffer if there is one available.
// The termination character is disabled after having been written to the
// buffer. Returns p to allow chaining of write commands.
func (pbuf *printBuf) writeTerm() {
	if pbuf.tc != 0 {
		_, err := pbuf.tw.Write([]byte{pbuf.tc})
		pbuf.es.send(err)
		pbuf.tc = 0
	}
}
