package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// loweredLeadingThenSearch lowers a single-spell-ability card whose spell
// ability is a leading effect sequence followed by a trailing library search,
// returning the resolved instruction sequence. It fails the test on any
// diagnostic or on a shape other than at least one leading instruction followed
// by a trailing Search instruction.
func loweredLeadingThenSearch(t *testing.T, typeLine, oracleText string) []game.Instruction {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Leading Search Test",
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
		t.Fatalf("lowerExecutableFaces(%q) sequence = %#v, want leading plus search", oracleText, seq)
	}
	last := seq[len(seq)-1]
	if _, ok := last.Primitive.(game.Search); !ok {
		t.Fatalf("lowerExecutableFaces(%q) last primitive = %#v, want game.Search", oracleText, last.Primitive)
	}
	if _, ok := seq[0].Primitive.(game.Search); ok {
		t.Fatalf("lowerExecutableFaces(%q) first primitive is a Search; want a leading effect", oracleText)
	}
	return seq
}

func TestLowerLeadingSequenceThenSearch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
	}{
		{
			name:       "bounce then land search",
			typeLine:   "Instant",
			oracleText: "Return up to one target nonland permanent to its owner's hand. Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		},
		{
			name:       "gain life then land search",
			typeLine:   "Sorcery",
			oracleText: "You gain 3 life. Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			loweredLeadingThenSearch(t, test.typeLine, test.oracleText)
		})
	}
}

// TestLowerLeadingSequenceThenSearchFailsClosed guards the conservative limits
// of the leading-sequence-then-search fallback: it must not pick up a trailing
// search whose shape is unsupported (a multi-card battlefield tutor), a body
// gated by a leading condition, or a library-top tutor trailing a leading
// effect.
func TestLowerLeadingSequenceThenSearchFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
	}{
		{
			name:       "unsupported multi-card battlefield search",
			typeLine:   "Sorcery",
			oracleText: "Draw a card. Search your library for any number of basic land cards, put them onto the battlefield, then shuffle.",
		},
		{
			name:       "leading conditional gate",
			typeLine:   "Sorcery",
			oracleText: "If you control a Forest, you gain 3 life. Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		},
		{
			name:       "library-top tutor after leading effect",
			typeLine:   "Instant",
			oracleText: "You gain 3 life. Search your library for a card, then shuffle and put that card on top.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Leading Search Fail Test",
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
