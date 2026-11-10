// Package cardname derives ASCII Go source names (file names and exported
// identifiers) for authored cards from their exact printed names. The printed
// name itself, kept in CardFace.Name and the registry, is never altered: only
// the derived source names fold Latin diacritics ("Lim-Dûl's Vault" becomes
// lim_dul_s_vault.go and LimDulSVault) so they satisfy the repository's
// ^[_A-Za-z0-9][_A-Za-z0-9-]*\.go$ file-name rule.
package cardname

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// latinFold transliterates U+00C0 through U+017F by offset: a letter is its
// ASCII base, '*' marks a ligature listed in latinLigatures, and '.' marks a
// non-letter (× and ÷), which separates words like other punctuation.
const latinFold = "AAAAAA*CEEEEIIIIDNOOOOO.OUUUUY**aaaaaa*ceeeeiiiidnooooo.ouuuuy*y" + // U+00C0
	"AaAaAaCcCcCcCcDdDdEeEeEeEeEeGgGgGgGgHhHhIiIiIiIiIi**JjKkkLlLlLlL" + // U+0100
	"lLlNnNnNnnNnOoOoOo**RrRrRrSsSsSsSsTtTtTtUuUuUuUuUuUuWwYyYZzZzZzs" //   U+0140

const latinFoldStart = 0xC0

var latinLigatures = map[rune]string{
	'Æ': "AE", 'æ': "ae", 'Þ': "Th", 'þ': "th", 'ß': "ss",
	'Ĳ': "IJ", 'ĳ': "ij", 'Œ': "OE", 'œ': "oe",
}

// words splits a printed name into ASCII words. Every maximal run of letters
// and digits is one word; Latin letters with diacritics fold to their ASCII
// base. Any other letter or digit has no faithful ASCII spelling and is an
// error, so a derived source name never silently drops part of a card name.
func words(name string) ([]string, error) {
	var result []string
	var word strings.Builder
	flush := func() {
		if word.Len() > 0 {
			result = append(result, word.String())
			word.Reset()
		}
	}
	for _, r := range name {
		switch {
		case r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			_ = word.WriteByte(byte(r))
		case r >= latinFoldStart && int(r-latinFoldStart) < len(latinFold):
			switch folded := latinFold[r-latinFoldStart]; folded {
			case '.':
				flush()
			case '*':
				_, _ = word.WriteString(latinLigatures[r])
			default:
				_ = word.WriteByte(folded)
			}
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			return nil, fmt.Errorf("card name %q contains %q (%U) with no ASCII transliteration", name, r, r)
		default:
			flush()
		}
	}
	flush()
	if len(result) == 0 {
		return nil, fmt.Errorf("card name %q has no letters or digits", name)
	}
	return result, nil
}

// goFileSuffixes are trailing name words that the go tool would read as a
// GOOS/GOARCH build constraint; such a file name gets a "_card" suffix.
var goFileSuffixes = []string{
	"aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos", "ios",
	"js", "linux", "nacl", "netbsd", "openbsd", "plan9", "solaris", "wasip1",
	"windows", "zos",
	"386", "amd64", "amd64p32", "arm", "arm64", "arm64be", "armbe", "loong64",
	"mips", "mips64", "mips64le", "mips64p32", "mips64p32le", "mipsle", "ppc",
	"ppc64", "ppc64le", "riscv", "riscv64", "s390", "s390x", "sparc", "sparc64",
	"wasm",
}

// FileStem returns the lowercase snake_case ASCII base name (without ".go")
// for a card's source file, for example "el_hajjaj" for "El-Hajjâj". A stem
// that would be read as a registry file, test file or platform-constrained
// file is suffixed with "_card". Append "_test" for its scenario file.
func FileStem(name string) (string, error) {
	parts, err := words(name)
	if err != nil {
		return "", err
	}
	for i, part := range parts {
		parts[i] = strings.ToLower(part)
	}
	stem := strings.Join(parts, "_")
	last := parts[len(parts)-1]
	if stem == "cards" || last == "test" || slices.Contains(goFileSuffixes, last) {
		stem += "_card"
	}
	return stem, nil
}

// Symbol returns the exported ASCII Go identifier for a card, for example
// "ElHajjaj" for "El-Hajjâj": each word is capitalized and concatenated, and a
// name beginning with a digit is prefixed with "Card".
func Symbol(name string) (string, error) {
	parts, err := words(name)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if parts[0][0] >= '0' && parts[0][0] <= '9' {
		_, _ = b.WriteString("Card")
	}
	for _, part := range parts {
		_, _ = b.WriteString(strings.ToUpper(part[:1]))
		_, _ = b.WriteString(part[1:])
	}
	return b.String(), nil
}
