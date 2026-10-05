package output

import (
	"os"

	"golang.org/x/term"
)

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// IsTerminalStdout reports whether stdout is a TTY.
func IsTerminalStdout() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// IsTerminalStdin reports whether stdin is a TTY.
func IsTerminalStdin() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
