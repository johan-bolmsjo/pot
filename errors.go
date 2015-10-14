package pot

import "fmt"

// Parse error containing location information.
type ParseError struct {
	Location Location
	Message  string
}

// Implements error.
func (err *ParseError) Error() string {
	return fmt.Sprintf("%s: %s", &err.Location, err.Message)
}
