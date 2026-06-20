package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerOptSequencesSpellParagraphs(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Opt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Scry 1. (Look at the top card of your library. You may put that card on the bottom.)\nDraw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Opt did not lower to a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want scry then draw", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Scry); !ok {
		t.Fatalf("first primitive = %T, want Scry", mode.Sequence[0].Primitive)
	}
	draw, ok := mode.Sequence[1].Primitive.(game.Draw)
	if !ok || draw.Player.Kind() != game.PlayerReferenceController || draw.Amount.Value() != 1 {
		t.Fatalf("second primitive = %#v, want controller draw one", mode.Sequence[1].Primitive)
	}
}

func TestSequentialSpellParagraphsFailClosedWhenSuffixTargets(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Crooked Opt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Draw a card.\nDeal 2 damage to any target.",
	})
	if face.SpellAbility.Exists && len(face.SpellAbility.Val.Modes[0].Sequence) > 1 {
		t.Fatalf("targeted suffix paragraph unexpectedly merged: %#v", face.SpellAbility.Val)
	}
}
