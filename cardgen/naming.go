package cardgen

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
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

// varNameInitialisms lists whole card-name words that Go naming conventions
// render entirely in capitals. staticcheck's ST1003 and revive's var-naming rule
// both treat these as initialisms and reject the mixed-case spelling, so a card
// named "Ram Through" must generate `var RAMThrough`, not `var RamThrough`, for
// the generated identifier to be simultaneously byte-identical to a curated file
// and lint-clean. Matching happens on whole words only (CardNameToVarName splits
// the name on non-alphanumeric runes first), which mirrors the linters'
// word-boundary rule: the standalone token "Ram" is rewritten while "Ramunap",
// "Rampage", and "Bramble" are left untouched.
//
// Keep this set conservative. Add an entry only when it belongs to the linters'
// common-initialism sets AND appears as a standalone word in a card name, so the
// generator never diverges from what the linters demand. "SIP" ("Sip of
// Hemlock") is the next known candidate but is intentionally excluded here; its
// card is generated, so adding it belongs to its own change. See issue #2892.
var varNameInitialisms = map[string]bool{
	"RAM": true,
}

// CardNameToVarName converts a card name to a Go exported variable name. Each
// maximal run of letters and digits becomes one capitalized word; a word that
// spells a known initialism (see varNameInitialisms) is rendered in all capitals
// so generated identifiers match what staticcheck's ST1003 and revive's
// var-naming rule require. A name that begins with a digit run is prefixed with
// "Card" to keep the identifier a valid, exported Go name.
func CardNameToVarName(name string) string {
	var b strings.Builder
	appendWord := func(word string) {
		if word == "" {
			return
		}
		if b.Len() == 0 {
			if r, _ := utf8.DecodeRuneInString(word); unicode.IsDigit(r) {
				_, _ = b.WriteString("Card")
			}
		}
		if varNameInitialisms[strings.ToUpper(word)] {
			_, _ = b.WriteString(strings.ToUpper(word))
			return
		}
		r, size := utf8.DecodeRuneInString(word)
		_, _ = b.WriteRune(unicode.ToUpper(r))
		_, _ = b.WriteString(word[size:])
	}
	start := -1
	for i, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if start < 0 {
				start = i
			}
			continue
		}
		if start >= 0 {
			appendWord(name[start:i])
			start = -1
		}
	}
	if start >= 0 {
		appendWord(name[start:])
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

// CardNameToPackageLetter returns the base ASCII lowercase letter of the card
// name for use as its Go package (directory) name. Latin diacritics on the first
// letter are folded to their base a-z letter, so "Éomer" sorts into the ordinary
// "e" package rather than a non-ASCII package the toolchain rejects. Names whose
// first letter has no single ASCII base (non-Latin scripts) fall into "other".
func CardNameToPackageLetter(name string) string {
	for _, r := range name {
		if !unicode.IsLetter(r) {
			continue
		}
		lower := unicode.ToLower(r)
		if lower >= 'a' && lower <= 'z' {
			return string(lower)
		}
		if base, ok := latinASCIIFold[lower]; ok {
			return string(rune(base))
		}
		return "other"
	}
	return "other"
}

// latinASCIIFold maps lowercase Latin letters bearing diacritics to their base
// ASCII a-z letter. It only needs to cover first-letter diacritics that appear in
// card names (currently just "É"), but modeling the full common Latin range keeps
// package assignment stable for any future accented card without a special case.
var latinASCIIFold = map[rune]byte{
	'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a', 'ā': 'a', 'ă': 'a', 'ą': 'a', 'æ': 'a',
	'ç': 'c', 'ć': 'c', 'ĉ': 'c', 'ċ': 'c', 'č': 'c',
	'ď': 'd', 'đ': 'd',
	'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e', 'ē': 'e', 'ĕ': 'e', 'ė': 'e', 'ę': 'e', 'ě': 'e',
	'ĝ': 'g', 'ğ': 'g', 'ġ': 'g', 'ģ': 'g',
	'ĥ': 'h', 'ħ': 'h',
	'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i', 'ĩ': 'i', 'ī': 'i', 'ĭ': 'i', 'į': 'i', 'ı': 'i',
	'ĵ': 'j',
	'ķ': 'k',
	'ĺ': 'l', 'ļ': 'l', 'ľ': 'l', 'ł': 'l',
	'ñ': 'n', 'ń': 'n', 'ņ': 'n', 'ň': 'n',
	'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ø': 'o', 'ō': 'o', 'ŏ': 'o', 'ő': 'o', 'œ': 'o',
	'ŕ': 'r', 'ŗ': 'r', 'ř': 'r',
	'ś': 's', 'ŝ': 's', 'ş': 's', 'š': 's',
	'ţ': 't', 'ť': 't', 'ŧ': 't',
	'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u', 'ũ': 'u', 'ū': 'u', 'ŭ': 'u', 'ů': 'u', 'ű': 'u', 'ų': 'u',
	'ŵ': 'w',
	'ý': 'y', 'ÿ': 'y', 'ŷ': 'y',
	'ź': 'z', 'ż': 'z', 'ž': 'z',
}
