package deck_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/deck"
)

func findEntry(entries []deck.Entry, name string) (deck.Entry, bool) {
	for _, e := range entries {
		if e.Name == name {
			return e, true
		}
	}
	return deck.Entry{}, false
}

func mustParse(t *testing.T, text string) *deck.Decklist {
	t.Helper()
	d, err := deck.Parse(strings.NewReader(text))
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	return d
}

func TestParseWithCommanderHeader(t *testing.T) {
	text := `// Commander
1 Atraxa, Praetors' Voice

// Deck
1 Sol Ring
1 Arcane Signet
3 Forest
`
	d := mustParse(t, text)

	if len(d.Commander) != 1 {
		t.Fatalf("Commander = %v, want 1 entry", d.Commander)
	}
	if d.Commander[0] != (deck.Entry{Quantity: 1, Name: "Atraxa, Praetors' Voice"}) {
		t.Errorf("commander entry = %+v", d.Commander[0])
	}
	if len(d.Cards) != 3 {
		t.Fatalf("Cards = %v, want 3 entries", d.Cards)
	}
	if forest, ok := findEntry(d.Cards, "Forest"); !ok || forest.Quantity != 3 {
		t.Errorf("Forest entry = %+v, ok=%v", forest, ok)
	}
	if got, want := d.Count(), 6; got != want {
		t.Errorf("Count() = %d, want %d", got, want)
	}
}

func TestParseWithoutCommanderHeader(t *testing.T) {
	text := `1 Sol Ring
1 Arcane Signet
2 Island
`
	d := mustParse(t, text)

	if len(d.Commander) != 0 {
		t.Errorf("Commander = %v, want empty", d.Commander)
	}
	if len(d.Cards) != 3 {
		t.Fatalf("Cards = %v, want 3 entries", d.Cards)
	}
	if got, want := d.Count(), 4; got != want {
		t.Errorf("Count() = %d, want %d", got, want)
	}
}

func TestParseInlineCommander(t *testing.T) {
	text := `COMMANDER: Atraxa, Praetors' Voice
1 Sol Ring
`
	d := mustParse(t, text)

	if len(d.Commander) != 1 || d.Commander[0].Name != "Atraxa, Praetors' Voice" {
		t.Fatalf("Commander = %+v, want single Atraxa", d.Commander)
	}
	// The inline header must not make subsequent lines commanders.
	if len(d.Cards) != 1 || d.Cards[0].Name != "Sol Ring" {
		t.Fatalf("Cards = %+v, want single Sol Ring", d.Cards)
	}
}

func TestParseQuantityFormsAndAnnotations(t *testing.T) {
	text := `4x Lightning Bolt (2X2) 117
1 Sol Ring *F*
2 Llanowar Elves (M19)
`
	d := mustParse(t, text)

	want := map[string]int{
		"Lightning Bolt": 4,
		"Sol Ring":       1,
		"Llanowar Elves": 2,
	}
	if len(d.Cards) != len(want) {
		t.Fatalf("Cards = %+v, want %d entries", d.Cards, len(want))
	}
	for name, qty := range want {
		e, ok := findEntry(d.Cards, name)
		if !ok {
			t.Errorf("missing entry %q", name)
			continue
		}
		if e.Quantity != qty {
			t.Errorf("%q quantity = %d, want %d", name, e.Quantity, qty)
		}
	}
}

func TestParseSideboardIgnored(t *testing.T) {
	text := `1 Sol Ring
SB: 1 Pithing Needle
// Sideboard
1 Relic of Progenitus
`
	d := mustParse(t, text)

	if len(d.Cards) != 1 || d.Cards[0].Name != "Sol Ring" {
		t.Fatalf("Cards = %+v, want only Sol Ring", d.Cards)
	}
}

func TestParseCommentEndsCommanderSection(t *testing.T) {
	// A category comment after the commander section must revert to the main
	// deck, so Birds of Paradise is not treated as a commander.
	text := `// Commander
1 Atraxa, Praetors' Voice
// Creatures (1)
1 Birds of Paradise
`
	d := mustParse(t, text)

	if len(d.Commander) != 1 || d.Commander[0].Name != "Atraxa, Praetors' Voice" {
		t.Fatalf("Commander = %+v, want single Atraxa", d.Commander)
	}
	if len(d.Cards) != 1 || d.Cards[0].Name != "Birds of Paradise" {
		t.Fatalf("Cards = %+v, want single Birds of Paradise", d.Cards)
	}
}

func TestParseCommanderEndedByBlankLine(t *testing.T) {
	// A "// Commander" header with no later "// Deck" header: a blank line must
	// end the commander section so the deck is not absorbed into Commander.
	text := `// Commander
1 Atraxa, Praetors' Voice

1 Sol Ring
1 Arcane Signet
`
	d := mustParse(t, text)

	if len(d.Commander) != 1 || d.Commander[0].Name != "Atraxa, Praetors' Voice" {
		t.Fatalf("Commander = %+v, want single Atraxa", d.Commander)
	}
	if len(d.Cards) != 2 {
		t.Fatalf("Cards = %+v, want 2 entries", d.Cards)
	}
}

func TestParseCompanionIgnored(t *testing.T) {
	// Canonical Moxfield order: Commander / Companion / Deck. The companion is
	// outside the deck and must not inflate the main-deck count.
	text := `// Commander
1 Atraxa, Praetors' Voice
// Companion
1 Lurrus of the Dream-Den
// Deck
1 Sol Ring
`
	d := mustParse(t, text)

	if len(d.Commander) != 1 || d.Commander[0].Name != "Atraxa, Praetors' Voice" {
		t.Fatalf("Commander = %+v, want single Atraxa", d.Commander)
	}
	if len(d.Cards) != 1 || d.Cards[0].Name != "Sol Ring" {
		t.Fatalf("Cards = %+v, want only Sol Ring (companion ignored)", d.Cards)
	}
}

func TestParseSplitCardPreserved(t *testing.T) {
	d := mustParse(t, "1 Fire // Ice\n")
	if len(d.Cards) != 1 || d.Cards[0].Name != "Fire // Ice" {
		t.Fatalf("Cards = %+v, want Fire // Ice preserved", d.Cards)
	}
}

func TestParseRealParentheticalNamePreserved(t *testing.T) {
	d := mustParse(t, "1 Erase (Not the Urza's Legacy One)\n")
	if len(d.Cards) != 1 || d.Cards[0].Name != "Erase (Not the Urza's Legacy One)" {
		t.Fatalf("Cards = %+v, want parenthetical name preserved", d.Cards)
	}
}

func TestParseMalformedLines(t *testing.T) {
	text := `1 Sol Ring
notanumber Card
0 Bad Card
2 Good Card
`
	d, err := deck.Parse(strings.NewReader(text))
	if err == nil {
		t.Fatal("Parse returned nil error, want parse errors")
	}

	// Best effort: the well-formed lines are still parsed.
	if _, ok := findEntry(d.Cards, "Sol Ring"); !ok {
		t.Error("expected Sol Ring to parse despite later errors")
	}
	if _, ok := findEntry(d.Cards, "Good Card"); !ok {
		t.Error("expected Good Card to parse despite earlier errors")
	}

	perrs := parseErrorsOf(err)
	if len(perrs) != 2 {
		t.Fatalf("got %d parse errors, want 2: %v", len(perrs), err)
	}
	if perrs[0].Line != 2 || !strings.Contains(perrs[0].Text, "notanumber") {
		t.Errorf("first error = %+v, want line 2 about notanumber", perrs[0])
	}
	if perrs[1].Line != 3 || !strings.Contains(perrs[1].Reason, "positive") {
		t.Errorf("second error = %+v, want line 3 about positive quantity", perrs[1])
	}
}

func TestParseCRLF(t *testing.T) {
	d := mustParse(t, "1 Sol Ring\r\n2 Island\r\n")
	if len(d.Cards) != 2 {
		t.Fatalf("Cards = %+v, want 2 entries", d.Cards)
	}
	if sol, ok := findEntry(d.Cards, "Sol Ring"); !ok || sol.Quantity != 1 {
		t.Errorf("Sol Ring entry = %+v, ok=%v", sol, ok)
	}
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deck.txt")
	content := "// Commander\n1 Krenko, Mob Boss\n// Deck\n30 Mountain\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	d, err := deck.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(d.Commander) != 1 || d.Commander[0].Name != "Krenko, Mob Boss" {
		t.Errorf("Commander = %+v", d.Commander)
	}
	if mtn, ok := findEntry(d.Cards, "Mountain"); !ok || mtn.Quantity != 30 {
		t.Errorf("Mountain entry = %+v, ok=%v", mtn, ok)
	}
}

func TestParseFileMissing(t *testing.T) {
	if _, err := deck.ParseFile(filepath.Join(t.TempDir(), "nope.txt")); err == nil {
		t.Fatal("ParseFile of missing path returned nil error")
	}
}

// parseErrorsOf extracts every ParseError joined into err.
func parseErrorsOf(err error) []*deck.ParseError {
	var out []*deck.ParseError
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range joined.Unwrap() {
			var pe *deck.ParseError
			if errors.As(e, &pe) {
				out = append(out, pe)
			}
		}
		return out
	}
	var pe *deck.ParseError
	if errors.As(err, &pe) {
		out = append(out, pe)
	}
	return out
}
