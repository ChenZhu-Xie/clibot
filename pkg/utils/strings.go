package utils

import (
	"strings"
	"unicode/utf8"
)

// TruncateRuneSafe trims s to at most maxRunes Unicode code points and appends
// "..." if it was shortened. It also strips any invalid UTF-8 sequences.
func TruncateRuneSafe(s string, maxRunes int) string {
	// Strip invalid UTF-8 bytes
	s = strings.Map(func(r rune) rune {
		if r == utf8.RuneError {
			return -1 // drop replacement characters from bad sequences
		}
		return r
	}, s)
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) > maxRunes {
		if maxRunes <= 3 {
			return string(runes[:maxRunes])
		}
		return string(runes[:maxRunes-3]) + "..."
	}
	return s
}
