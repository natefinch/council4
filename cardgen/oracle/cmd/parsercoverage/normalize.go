package main

import (
	"strings"
	"unicode"
)

// normalizeCluster collapses an uncovered span's text into a stable cluster key:
// it lowercases, collapses runs of whitespace, and replaces bare integers with N
// so that grammar that differs only by a numeric literal clusters together.
func normalizeCluster(text string) string {
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
