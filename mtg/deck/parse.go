package deck

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ParseError describes a single decklist line that could not be parsed.
type ParseError struct {
	// Line is the 1-based line number of the offending line.
	Line int

	// Text is the offending line as written, before trimming.
	Text string

	// Reason explains why the line could not be parsed.
	Reason string
}

// Error implements error.
func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s: %q", e.Line, e.Reason, e.Text)
}

// section identifies which part of the decklist subsequent entries belong to.
type section int

const (
	sectionMain section = iota
	sectionCommander
	sectionIgnore
)

var (
	// setCollectorRe matches a trailing set code plus collector number, such as
	// " (2X2) 117". The set code may be any case; the collector is the final
	// whitespace-delimited token.
	setCollectorRe = regexp.MustCompile(`\s+\([0-9A-Za-z]{2,6}\)\s+\S+$`)

	// setOnlyRe matches a trailing uppercase set code with no collector number,
	// such as " (C21)". Uppercase-only avoids stripping real parenthetical card
	// names like "Erase (Not the Urza's Legacy One)".
	setOnlyRe = regexp.MustCompile(`\s+\([0-9A-Z]{2,5}\)$`)

	// foilRe matches a trailing foil/etched marker such as " *F*" or " *E*".
	foilRe = regexp.MustCompile(`\s+\*[A-Za-z]\*$`)
)

// ParseFile reads and parses the decklist at path.
func ParseFile(path string) (*Decklist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return Parse(f)
}

// Parse reads a Commander decklist in Moxfield/MTGO text format from r.
//
// It recognizes "N Card Name" and "Nx Card Name" entries, a commander section
// introduced by a "// Commander" header or a "COMMANDER:" line, and ignores
// blank lines, comments, and sideboard lines ("SB:" or a "// Sideboard"
// header). Trailing set/collector and foil annotations are trimmed from names.
//
// Parse always returns a best-effort Decklist. When one or more lines cannot be
// parsed, it skips them and also returns an error joining every ParseError;
// callers can use errors.As to inspect individual line failures.
func Parse(r io.Reader) (*Decklist, error) {
	d := &Decklist{}
	var parseErrs []error
	current := sectionMain

	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		raw := strings.TrimPrefix(scanner.Text(), "\ufeff")
		line := strings.TrimSpace(raw)
		if line == "" {
			// A blank line ends the (short) commander section so a header
			// layout without an explicit "// Deck" does not absorb the deck.
			if current == sectionCommander {
				current = sectionMain
			}
			continue
		}

		if h, ok := parseHeader(line); ok {
			if h.inlineName == "" {
				current = h.section
				continue
			}
			if entry, ok := parseEntryLoose(h.inlineName); ok {
				d.add(h.section, entry)
			}
			continue
		}

		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			// An unrecognized comment ends the commander section: category
			// comments such as "// Creatures (30)" belong to the main deck.
			if current == sectionCommander {
				current = sectionMain
			}
			continue
		}

		if current == sectionIgnore {
			continue
		}

		entry, perr := parseEntry(line, lineNum, raw)
		if perr != nil {
			parseErrs = append(parseErrs, perr)
			continue
		}
		d.add(current, entry)
	}
	if err := scanner.Err(); err != nil {
		return d, err
	}
	if len(parseErrs) > 0 {
		return d, errors.Join(parseErrs...)
	}
	return d, nil
}

func (d *Decklist) add(sec section, e Entry) {
	switch sec {
	case sectionCommander:
		d.Commander = append(d.Commander, e)
	case sectionMain:
		d.Cards = append(d.Cards, e)
	default:
	}
}

// header is a recognized section header, optionally carrying an inline entry
// such as the "Atraxa, Praetors' Voice" in "COMMANDER: Atraxa, Praetors' Voice".
type header struct {
	section    section
	inlineName string
}

// parseHeader recognizes a "//"-style header ("// Commander", "// Deck",
// "// Sideboard") or a "Keyword:" line ("COMMANDER: ...", "SB: ..."). An inline
// entry after a "Keyword:" header is a one-off and does not change the persistent
// section.
func parseHeader(line string) (header, bool) {
	if rest, ok := strings.CutPrefix(line, "//"); ok {
		if sec, ok := sectionForKeyword(normalizeHeaderLabel(rest)); ok {
			return header{section: sec}, true
		}
		return header{}, false
	}
	idx := strings.IndexByte(line, ':')
	if idx <= 0 {
		return header{}, false
	}
	sec, ok := sectionForKeyword(normalizeHeaderLabel(line[:idx]))
	if !ok {
		return header{}, false
	}
	return header{section: sec, inlineName: strings.TrimSpace(line[idx+1:])}, true
}

// normalizeHeaderLabel reduces a header label to a bare lowercase keyword,
// stripping a trailing "(N)" count and ":" suffix.
func normalizeHeaderLabel(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndexByte(s, '('); i >= 0 && strings.HasSuffix(s, ")") {
		s = strings.TrimSpace(s[:i])
	}
	s = strings.TrimSuffix(s, ":")
	return strings.ToLower(strings.TrimSpace(s))
}

func sectionForKeyword(kw string) (section, bool) {
	switch kw {
	case "commander", "commanders":
		return sectionCommander, true
	case "deck", "mainboard", "maindeck", "main":
		return sectionMain, true
	case "sideboard", "maybeboard", "companion", "sb":
		return sectionIgnore, true
	}
	return sectionMain, false
}

// parseEntry parses a single "N Name" / "Nx Name" entry. lineNum and raw are
// used to build a ParseError on failure.
func parseEntry(s string, lineNum int, raw string) (Entry, *ParseError) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return Entry{}, &ParseError{Line: lineNum, Text: raw, Reason: "empty entry"}
	}
	qty, ok := parseQuantity(fields[0])
	if !ok {
		return Entry{}, &ParseError{Line: lineNum, Text: raw, Reason: "missing or invalid quantity"}
	}
	if qty <= 0 {
		return Entry{}, &ParseError{Line: lineNum, Text: raw, Reason: "quantity must be positive"}
	}
	name := cleanName(strings.TrimSpace(s[len(fields[0]):]))
	if name == "" {
		return Entry{}, &ParseError{Line: lineNum, Text: raw, Reason: "missing card name"}
	}
	return Entry{Quantity: qty, Name: name}, nil
}

// parseEntryLoose parses an inline entry that may omit its quantity, such as the
// bare card name in "COMMANDER: Atraxa, Praetors' Voice" (quantity defaults to
// 1). It reports ok=false for an empty or invalid entry.
func parseEntryLoose(s string) (Entry, bool) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return Entry{}, false
	}
	if qty, ok := parseQuantity(fields[0]); ok {
		if qty <= 0 {
			return Entry{}, false
		}
		name := cleanName(strings.TrimSpace(s[len(fields[0]):]))
		if name == "" {
			return Entry{}, false
		}
		return Entry{Quantity: qty, Name: name}, true
	}
	name := cleanName(strings.TrimSpace(s))
	if name == "" {
		return Entry{}, false
	}
	return Entry{Quantity: 1, Name: name}, true
}

// parseQuantity parses a leading quantity token, accepting both "4" and "4x".
func parseQuantity(tok string) (int, bool) {
	tok = strings.TrimSuffix(tok, "x")
	tok = strings.TrimSuffix(tok, "X")
	n, err := strconv.Atoi(tok)
	if err != nil {
		return 0, false
	}
	return n, true
}

// cleanName strips trailing foil and set/collector annotations from a card name.
func cleanName(name string) string {
	name = foilRe.ReplaceAllString(name, "")
	name = setCollectorRe.ReplaceAllString(name, "")
	name = setOnlyRe.ReplaceAllString(name, "")
	return strings.TrimSpace(name)
}
