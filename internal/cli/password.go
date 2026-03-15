package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PromptPassword reads a password from the terminal without echoing.
// Falls back to plain stdin reading if terminal is not available (e.g., piped input).
func PromptPassword() (string, error) {
	fmt.Fprint(os.Stderr, "Password: ")

	// Try terminal masked input first
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		passwordBytes, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr) // newline after masked input
		if err != nil {
			return "", err
		}
		return string(passwordBytes), nil
	}

	// Fallback: read from stdin (piped input)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
