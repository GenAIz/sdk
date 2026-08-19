package dialogz

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ConfirmYes is an io.Writer/io.reader dialog confirmation routine which expects a Yes/No or y/n answer from the user.
// Spaces and letter casing in the answer are ignored. The routine expects a valid character to determine if Yes is true
// or false, followed by '\n'.
//
// The routine will keep asking the user with the provided message until it gets a valid answer.
func ConfirmYes(w io.Writer, r io.Reader, message string) bool {
	var reader = bufio.NewReader(r)

	for {
		if _, err := fmt.Fprintf(w, "%s (%s) ", message, "[y/n]"); err == nil {
			var s, _ = reader.ReadString('\n')

			s = strings.TrimSpace(s)
			s = strings.ToLower(s)

			if s != "" {
				if s == "y" || s == "yes" {
					return true
				}

				if s == "n" || s == "no" {
					return false
				}
			}
		} else {
			return false
		}
	}
}
