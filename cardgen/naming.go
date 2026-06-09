package cardgen

import (
	"strings"
	"unicode"
)

// CardNameToVarName converts a card name to a Go exported variable name.
func CardNameToVarName(name string) string {
	var b strings.Builder
	capitalize := true
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalize = true
			continue
		}
		if capitalize {
			_, _ = b.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else {
			_, _ = b.WriteRune(r)
		}
	}
	return b.String()
}

// CardNameToFileName converts a card name to a snake_case file name.
func CardNameToFileName(name string) string {
	var b strings.Builder
	prevWasUnderscore := false
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			if !prevWasUnderscore && b.Len() > 0 {
				_, _ = b.WriteRune('_')
				prevWasUnderscore = true
			}
			continue
		}
		_, _ = b.WriteRune(unicode.ToLower(r))
		prevWasUnderscore = false
	}
	return strings.TrimSuffix(b.String(), "_")
}

// CardNameToPackageLetter returns the lowercase first letter of the card name.
func CardNameToPackageLetter(name string) string {
	for _, r := range name {
		if unicode.IsLetter(r) {
			return string(unicode.ToLower(r))
		}
	}
	return "other"
}
