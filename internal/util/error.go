package util

import (
	"fmt"
)

type Error struct {
	Source string
	Info   string
	Err    error
}

// Converts the Error into a string containing all information about
// The error
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s %s %s", e.Source, e.Info, e.Err.Error())
	}
	return fmt.Sprintf("%s %s", e.Source, e.Info)
}

// Converts the Error into a short string containing only the brief info
func (e *Error) Short() string {
	return e.Info
}

// Same as Error call, just satisfies the Stringer interface
func (e *Error) String() string {
	return e.Error()
}
