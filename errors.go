package pot

import "fmt"

// Parse error containing location information and optional identifier (file name or similar).
type ParseError struct {
	Identifier string
	Location   Location
	Message    string
}

// Implements error.
func (err *ParseError) Error() string {
	if err.Identifier != "" {
		return fmt.Sprintf("%s:%s: %s", err.Identifier, &err.Location, err.Message)
	}
	return fmt.Sprintf("%s: %s", &err.Location, err.Message)
}
