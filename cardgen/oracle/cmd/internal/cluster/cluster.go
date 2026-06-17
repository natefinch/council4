// Package cluster normalizes uncovered Oracle-text spans into stable cluster
// keys. The parser-coverage and card-backlog tools both group unrepresented
// grammar by these keys so that spans differing only by whitespace or a numeric
// literal collapse into one ranked work-queue row.
package cluster

import (
	"strings"
	"unicode"
)

// Normalize collapses an uncovered span's text into a stable cluster key: it
// lowercases, collapses runs of whitespace, and replaces bare integers with N so
// that grammar that differs only by a numeric literal clusters together.
func Normalize(text string) string {
	fields := strings.Fields(strings.ToLower(text))
	for i := range fields {
		if isInteger(fields[i]) {
			fields[i] = "N"
		}
	}
	if len(fields) == 0 {
		return strings.TrimSpace(text)
	}
	return strings.Join(fields, " ")
}

func isInteger(token string) bool {
	if token == "" {
		return false
	}
	for _, r := range token {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
