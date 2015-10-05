package pot

// Error sink used to implement "abort on first error" functionality with ease.
// A function would check if the sink already contains an error and if it does
// do nothing.
type errorSink struct {
	err error
}

// Returns the error stored in the sink.
func (p *errorSink) error() error {
	return p.err
}

// Returns true if there is no error in the sink.
func (p *errorSink) ok() bool {
	if p.err == nil {
		return true
	}
	return false
}

// Sends an error to the sink.
// Does nothing if an error is already stored in the sink.
func (p *errorSink) send(err error) {
	if p.err == nil {
		p.err = err
	}
}
