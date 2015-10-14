package pot

import "fmt"

// Parser location in text input.
type Location struct {
	Line   int // Line number counting from zero.
	Column int // Column number counting from zero.
}

// Implements fmt.Stringer
func (location *Location) String() string {
	// For presentation the line number is adjusted to count from one.
	return fmt.Sprintf("%d:%d", location.Line+1, location.Column)
}
