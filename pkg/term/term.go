//go:build !windows

package term

import (
	"syscall"

	"golang.org/x/term"
)

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal() bool {
	return term.IsTerminal(syscall.Stdin)
}
