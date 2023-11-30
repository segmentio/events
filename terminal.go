package events

import "golang.org/x/term"

// IsTerminal returns true if the given file descriptor is a terminal.
//
// Deprecated: Use golang.org/x/term.IsTerminal instead.
func IsTerminal(fd int) bool {
	return term.IsTerminal(fd)
}
