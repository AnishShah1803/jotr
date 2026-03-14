package utils

import (
	"bufio"
	"os"
)

// StdinReader is an interface for reading from stdin.
// This allows for testable input operations by injecting mock readers.
type StdinReader interface {
	ReadString(delim byte) (string, error)
}

// OsStdinReader implements StdinReader using os.Stdin.
type OsStdinReader struct{}

// ReadString reads from os.Stdin until the delimiter is encountered.
func (r OsStdinReader) ReadString(delim byte) (string, error) {
	return bufio.NewReader(os.Stdin).ReadString(delim)
}

// DefaultStdinReader is the production reader using stdin.
var DefaultStdinReader StdinReader = OsStdinReader{}
