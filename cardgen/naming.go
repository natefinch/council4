package cardgen

import (
	"slices"
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

// CardNameToSafeFileName converts a card name to a file name that cannot be
// mistaken for a package registry, test file, or platform-specific Go file.
func CardNameToSafeFileName(name string) string {
	base := CardNameToFileName(name)
	if base == "cards" || strings.HasSuffix(base, "_test") {
		return base + "_card"
	}
	parts := strings.Split(base, "_")
	for _, suffix := range goFileSuffixes {
		if len(parts) >= len(suffix) && slices.Equal(parts[len(parts)-len(suffix):], suffix) {
			return base + "_card"
		}
	}
	return base
}

var goFileSuffixes = [][]string{
	{"aix"}, {"android"}, {"darwin"}, {"dragonfly"}, {"freebsd"}, {"illumos"},
	{"ios"}, {"js"}, {"linux"}, {"netbsd"}, {"openbsd"}, {"plan9"}, {"solaris"},
	{"wasip1"}, {"windows"},
	{"386"}, {"amd64"}, {"arm"}, {"arm64"}, {"loong64"}, {"mips"}, {"mips64"},
	{"mips64le"}, {"mipsle"}, {"ppc64"}, {"ppc64le"}, {"riscv64"}, {"s390x"},
	{"wasm"},
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
