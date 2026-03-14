package bot

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/keepmind9/clibot/pkg/constants"
)

// maskSecret masks sensitive information for logging
func maskSecret(s string) string {
	if len(s) <= constants.MinSecretLengthForMasking {
		return "***"
	}
	return s[:constants.SecretMaskPrefixLength] + "***" + s[len(s)-constants.SecretMaskSuffixLength:]
}

// sanitizeTokenFromError removes bot tokens from error messages to prevent
// accidental exposure in logs. Go's net/http includes the full URL in HTTP
// errors, which contains the bot token for Telegram API calls.
func sanitizeTokenFromError(token string, err error) error {
	if err == nil || token == "" {
		return err
	}
	msg := err.Error()
	sanitized := strings.ReplaceAll(msg, token, "<REDACTED>")
	if sanitized != msg {
		return fmt.Errorf("%s", sanitized)
	}
	return err
}

// WrapTablesInCodeBlocks detects markdown tables and wraps them in code blocks
// if they are not already inside one. This helps with mobile rendering by
// using fixed-width fonts.
func WrapTablesInCodeBlocks(text string) string {
	// Simple regex to detect markdown table header separator: | --- | or |---|
	// This is a common pattern for markdown tables.
	tableRegex := regexp.MustCompile(`(?m)^(\|?\s*:?-+:?\s*\|?)+\s*$`)
	lines := strings.Split(text, "\n")

	var result []string
	inCodeBlock := false
	inTable := false
	tableStart := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track code blocks to avoid double-wrapping
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}

		if inCodeBlock {
			result = append(result, line)
			continue
		}

		// Detect table separator
		if tableRegex.MatchString(trimmed) {
			if !inTable && i > 0 {
				// We found a separator, the previous line was likely the header
				inTable = true
				tableStart = len(result) - 1
				// Insert opening backticks before header
				if tableStart >= 0 {
					header := result[tableStart]
					result[tableStart] = "```\n" + header
				}
			}
		} else if inTable {
			// Check if table ended (empty line or doesn't start/contain |)
			if trimmed == "" || (!strings.Contains(trimmed, "|") && !tableRegex.MatchString(trimmed)) {
				inTable = false
				// Close the code block
				if len(result) > 0 {
					result[len(result)-1] = result[len(result)-1] + "\n```"
				}
			}
		}

		result = append(result, line)
	}

	// Close table if it reached the end of text
	if inTable && len(result) > 0 {
		result[len(result)-1] = result[len(result)-1] + "\n```"
	}

	return strings.Join(result, "\n")
}

// SplitMessage splits a string into chunks of at most maxSize.
// It prefers splitting at newlines, but falls back to safe rune boundaries
// if no newline is found within the limit.
func SplitMessage(s string, maxSize int) []string {
	if len(s) == 0 {
		return []string{""}
	}
	var chunks []string
	for len(s) > 0 {
		if len(s) <= maxSize {
			chunks = append(chunks, s)
			break
		}

		splitIdx := strings.LastIndex(s[:maxSize], "\n")
		// Only split at newline if it's reasonably close to the limit (e.g. > maxSize/2)
		// Otherwise, fall back to safe rune boundaries
		if splitIdx > maxSize/2 {
			splitIdx += 1 // Include the newline in the current chunk
		} else {
			splitIdx = findSafeRuneSplit(s, maxSize)
		}

		chunks = append(chunks, s[:splitIdx])
		s = s[splitIdx:]
	}
	return chunks
}

func findSafeRuneSplit(s string, maxSize int) int {
	if len(s) <= maxSize {
		return len(s)
	}
	idx := maxSize
	for idx > 0 && !utf8.RuneStart(s[idx]) {
		idx--
	}
	return idx
}
