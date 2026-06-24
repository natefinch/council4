package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// loweredSearchThenTrailing lowers a single-spell-ability card whose spell
// ability is a library search followed by trailing effects, returning the
// resolved instruction sequence. It fails the test on any diagnostic or on a
// shape other than a Search instruction followed by at least one trailing
// instruction.
func loweredSearchThenTrailing(t *testing.T, typeLine, oracleText string) []game.Instruction {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Trailing Search Test",
		Layout:     "normal",
		TypeLine:   typeLine,
		OracleText: oracleText,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("lowerExecutableFaces(%q) diagnostics = %#v", oracleText, diagnostics)
	}
	if len(faces) != 1 || !faces[0].SpellAbility.Exists {
		t.Fatalf("lowerExecutableFaces(%q) faces = %#v", oracleText, faces)
	}
	modes := faces[0].SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("lowerExecutableFaces(%q) modes = %#v", oracleText, modes)
	}
	seq := modes[0].Sequence
	if len(seq) < 2 {
		t.Fatalf("lowerExecutableFaces(%q) sequence = %#v, want search plus trailing", oracleText, seq)
	}
	if _, ok := seq[0].Primitive.(game.Search); !ok {
		t.Fatalf("lowerExecutableFaces(%q) first primitive = %#v, want game.Search", oracleText, seq[0].Primitive)
	}
	return seq
}

func TestLowerSearchThenTrailingSequence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		trailing   game.Primitive
	}{
		{
			name:       "search then create token",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a basic land card, reveal it, put it into your hand, then shuffle. Create a Food token. (It's an artifact with \"{2}, {T}, Sacrifice this token: You gain 3 life.\")",
		},
		{
			name:       "search then investigate",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Investigate. (Create a Clue token. It's an artifact with \"{2}, Sacrifice this token: Draw a card.\")",
		},
		{
			name:       "search then kicker-gated tokens",
			typeLine:   "Sorcery",
			oracleText: "Kicker {1}{W} (You may pay an additional {1}{W} as you cast this spell.)\nSearch your library for a basic land card, put it onto the battlefield tapped, then shuffle. If this spell was kicked, create two 1/1 white Soldier creature tokens.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			loweredSearchThenTrailing(t, test.typeLine, test.oracleText)
		})
	}
}

// TestLowerSearchThenTrailingSequenceFailsClosed guards the conservative limits
// of the trailing-search-sequence fallback: it must not pick up a search whose
// trailing sentence refers back to the found card, an activated-ability land
// search with a trailing effect, or a library-top tutor with a trailing effect.
func TestLowerSearchThenTrailingSequenceFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
	}{
		{
			name:       "trailing reference to found card",
			typeLine:   "Sorcery",
			oracleText: "Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Create a 0/1 colorless Eldrazi Spawn creature token. It has \"Sacrifice this token: Add {C}.\"",
		},
		{
			name:       "activated land search with trailing effect",
			typeLine:   "Land",
			oracleText: "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap target land.",
		},
		{
			name:       "library-top tutor with trailing effect",
			typeLine:   "Instant",
			oracleText: "Search your library for a card, then shuffle and put that card on top. Draw a card.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Trailing Search Fail Test",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("lowerExecutableFaces(%q) lowered; want fail closed", test.oracleText)
			}
		})
	}
}
