package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func franticSearchCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Frantic Search",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Instant",
		OracleText: "Draw two cards, then discard two cards. Untap up to three lands.",
		Colors:     []string{"U"},
	}
}

func TestLowerFranticSearchOrderedSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, franticSearchCard())
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %#v, want untargeted three-instruction sequence", mode)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("instruction 0 = %#v, want draw", mode.Sequence[0])
	}
	if _, ok := mode.Sequence[1].Primitive.(game.Discard); !ok {
		t.Fatalf("instruction 1 = %#v, want discard", mode.Sequence[1])
	}
	untap, ok := mode.Sequence[2].Primitive.(game.Untap)
	if !ok ||
		!untap.ChooseUpTo ||
		untap.Amount != game.Fixed(3) ||
		untap.Group.Domain() != game.GroupDomainBattlefield {
		t.Fatalf("instruction 2 = %#v, want choose up to three battlefield lands", mode.Sequence[2])
	}
	selection := untap.Group.Selection()
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Land {
		t.Fatalf("untap selection = %#v, want lands", selection)
	}
}

func TestGenerateFranticSearchExecutableSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(franticSearchCard(), "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Draw",
		"game.Discard",
		"game.Untap",
		"ChooseUpTo: true",
		"game.Fixed(3)",
		"types.Land",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerFranticSearchUntapNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Draw two cards, then discard two cards, then untap up to two lands.",
		"Draw two cards, then discard two cards, then untap up to three creatures.",
		"Draw two cards, then discard two cards, then untap up to three lands you control.",
		"Draw two cards, then discard two cards, then untap up to three random lands.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			card := franticSearchCard()
			card.OracleText = oracle
			_, diagnostics := lowerExecutableFaces(card)
			if len(diagnostics) == 0 {
				t.Fatalf("expected fail-closed diagnostics for %q", oracle)
			}
		})
	}
}
