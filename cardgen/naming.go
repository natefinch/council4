package cardgen

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

// GeneratedCardIdentity describes the Go identity and location of one generated
// card definition.
type GeneratedCardIdentity struct {
	RelativePath     string
	PackageName      string
	VariableName     string
	IdentifierSuffix string
	SupersededPath   string
}

// GeneratedIdentity returns the deterministic generated identity for a card.
// Playable tokens always live in their own namespace and include their complete
// Oracle UUID. Other cards use the ordinary letter package and are suffixed
// only when disambiguate is true.
func GeneratedIdentity(card *ScryfallCard, disambiguate bool) (GeneratedCardIdentity, error) {
	letter := CardNameToPackageLetter(card.Name)
	base := CardNameToSafeFileName(card.Name)
	identity := GeneratedCardIdentity{
		RelativePath: filepath.Join(letter, base+".go"),
		PackageName:  letter,
		VariableName: CardNameToVarName(card.Name),
	}
	if card.Layout == "token" || card.Layout == "double_faced_token" {
		oracleID, err := normalizedOracleUUID(card.OracleID)
		if err != nil {
			return GeneratedCardIdentity{}, fmt.Errorf("token %q: %w", card.Name, err)
		}
		identity.IdentifierSuffix = "Token" + oracleID
		identity.RelativePath = filepath.Join("tokens", letter, base+"_"+oracleID+".go")
		identity.VariableName += identity.IdentifierSuffix
		identity.SupersededPath = filepath.Join(letter, base+".go")
		return identity, nil
	}
	if !disambiguate {
		return identity, nil
	}
	identity.IdentifierSuffix = CardDisambiguationSuffix(card)
	if identity.IdentifierSuffix == "" {
		return GeneratedCardIdentity{}, fmt.Errorf("card %q has no Oracle or Scryfall ID", card.Name)
	}
	identity.RelativePath = filepath.Join(letter, base+"_"+strings.ToLower(identity.IdentifierSuffix)+".go")
	identity.VariableName += identity.IdentifierSuffix
	identity.SupersededPath = filepath.Join(letter, base+".go")
	return identity, nil
}

func normalizedOracleUUID(id string) (string, error) {
	parts := strings.Split(strings.ToLower(id), "-")
	wantedLengths := [...]int{8, 4, 4, 4, 12}
	if len(parts) != len(wantedLengths) {
		return "", fmt.Errorf("oracle ID %q is not a UUID", id)
	}
	for index, part := range parts {
		if len(part) != wantedLengths[index] {
			return "", fmt.Errorf("oracle ID %q is not a UUID", id)
		}
		for _, r := range part {
			if !strings.ContainsRune("0123456789abcdef", r) {
				return "", fmt.Errorf("oracle ID %q is not a UUID", id)
			}
		}
	}
	return strings.Join(parts, ""), nil
}

// CardNameToVarName converts a card name to a Go exported variable name.
func CardNameToVarName(name string) string {
	var b strings.Builder
	capitalize := true
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalize = true
			continue
		}
		if b.Len() == 0 && unicode.IsDigit(r) {
			_, _ = b.WriteString("Card")
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

// CardDisambiguationSuffix returns a stable Go identifier suffix derived from
// the card's Oracle identity, falling back to its Scryfall printing identity.
func CardDisambiguationSuffix(card *ScryfallCard) string {
	identity := card.OracleID
	if identity == "" {
		identity = card.ID
	}
	var b strings.Builder
	_, _ = b.WriteString("Scryfall")
	for _, r := range identity {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			_, _ = b.WriteRune(r)
		}
	}
	if b.Len() == len("Scryfall") {
		return ""
	}
	return b.String()
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
