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

func TestLowerFranticSearchBoundedUntapBroadForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		oracle  string
		amount  game.Quantity
		reqType types.Card
	}{
		{"Draw two cards, then discard two cards. Untap up to two lands.", game.Fixed(2), types.Land},
		{"Draw two cards, then discard two cards. Untap up to three creatures.", game.Fixed(3), types.Creature},
		{"Draw two cards, then discard two cards. Untap up to three lands you control.", game.Fixed(3), types.Land},
	}
	for _, tc := range cases {
		t.Run(tc.oracle, func(t *testing.T) {
			t.Parallel()
			card := franticSearchCard()
			card.OracleText = tc.oracle
			faces, diagnostics := lowerExecutableFaces(card)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			mode := faces[0].SpellAbility.Val.Modes[0]
			untap, ok := mode.Sequence[len(mode.Sequence)-1].Primitive.(game.Untap)
			if !ok ||
				!untap.ChooseUpTo ||
				untap.Amount != tc.amount ||
				untap.Group.Domain() != game.GroupDomainBattlefield {
				t.Fatalf("last instruction = %#v, want choose up to %v", mode.Sequence, tc.amount)
			}
			selection := untap.Group.Selection()
			if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != tc.reqType {
				t.Fatalf("untap selection = %#v, want %v", selection, tc.reqType)
			}
		})
	}
}

func TestLowerFranticSearchUntapNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
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
