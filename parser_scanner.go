package pot

// Wraps a parser to provide a bufio.Scanner like API for easy parsing.
type ParserScanner struct {
	parser    Parser
	subparser Parser
	es        *errorSink
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
	err := scanner.es.err()
	if err == nil {
		scanner.subparser, err = scanner.parser.Next()
		scanner.es.send(err)
	}
	if err != nil || scanner.subparser == nil {
		return false
	}
	return true
}

// Returns the previously scanned sub parser.
func (scanner *ParserScanner) SubParser() Parser {
	if scanner.es.ok() {
		return scanner.subparser
	}
	return nil
}

// Returns the first error that occured while scanning.
// This should be called after Scan() has returned false to check
// for errors.
func (scanner *ParserScanner) Err() error {
	return scanner.es.err()
}

// Inject an error into the scanner.
// This will cause the scanner to abort parser iteration and the injected error
// will be returned by the Err method.
func (scanner *ParserScanner) InjectError(err error) {
	scanner.es.send(err)
}
