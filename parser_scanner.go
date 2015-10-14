package pot

// Wraps a parser to provide a bufio.Scanner like API for easy parsing.
type ParserScanner struct {
	parser Parser
	result Parser
	es     *errorSink
}

// Create a new scanner operating on parser.
func NewParserScanner(parser Parser) *ParserScanner {
	return &ParserScanner{parser, nil, new(errorSink)}
}

// Create a new scanner operating on parser.
// Errors are sent to the specified error sink.
func newParserScannerErrorSink(parser Parser, es *errorSink) *ParserScanner {
	return &ParserScanner{parser, nil, es}
}

// Scan the parser for a sub parser.
// Returns true if a sub parser was found.
func (scanner *ParserScanner) Scan() bool {
	var err error
	if scanner.es.ok() {
		scanner.result, err = scanner.parser.Next()
		scanner.es.send(err)
	}
	if err != nil || scanner.result == nil {
		return false
	}
	return true
}

// Returns the result of the previous scan.
func (scanner *ParserScanner) Result() Parser {
	if scanner.es.ok() {
		return scanner.result
	}
	return nil
}

// Returns the first error that occured while scanning.
// This should be called after Scan() has returned false to check
// for errors.
func (scanner *ParserScanner) Error() error {
	return scanner.es.error()
}
