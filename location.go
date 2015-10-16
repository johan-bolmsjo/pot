package pot

import (
	"fmt"
	"unicode/utf8"
)

// Parser location in text input.
type Location struct {
	Line   uint32 // Line number counting from zero.
	Column uint32 // Column number counting from zero.
}

// Update location information (counting lines and columns) from a byte slice.
func (location *Location) updateFromBytes(bytes []byte) {
	for _, c := range bytes {
		switch {
		case c == '\r':
			location.Column = 0
		case c == '\n':
			location.Column = 0
			location.Line++
		case utf8.RuneStart(c):
			location.Column++
		}
	}
}

// Implements fmt.Stringer
func (location Location) String() string {
	// The line number is adjusted to count from one for presentation.
	return fmt.Sprintf("%d:%d", location.Line+1, location.Column)
}

// Format an error with the parser location in text input.
func (location Location) Errorf(format string, a ...interface{}) *ParseError {
	return &ParseError{
		Location: location,
		Message:  fmt.Sprintf(format, a...),
	}
}
