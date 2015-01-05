package main

import (
	"fmt"
)

// Error message indented both short abbreviated error message for
// client responses, and longer messages for error logging.
type ErroMsg struct {
	// Where the Error occurred at
	Source string
	// Short blurb about the error that will be sent back to the client
	Info string
	// Longer full error message
	Err error
}

// Converts the Error into a string containing all information about
// The error
func (e *ErroMsg) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s %s %s", e.Source, e.Info, e.Err.Error())
	}
	return fmt.Sprintf("%s %s", e.Source, e.Info)
}

// Converts the Error into a short string containing only the brief info
func (e *ErroMsg) Short() string {
	return e.Info
}

// Same as Error call, just satisfies the Stringer interface
func (e *ErroMsg) String() string {
	return e.Error()
}
