package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestPublishedReturnModalConsequences(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Modal Published Subject", Layout: "normal", TypeLine: "Instant",
		OracleText: "Choose one —\n• Exile target creature you control, then return it to the battlefield tapped under its owner's control. If it's an Elf, untap it.\n• Draw a card.",
	}
	face := lowerSingleFace(t, card)
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	put, ok := sequence[1].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatal("modal return did not lower to a battlefield entry")
	}
	untap, ok := sequence[2].Primitive.(game.Untap)
	if !ok {
		t.Fatal("modal consequence did not lower to untap")
	}
	want := game.LinkedObjectReference(string(put.PublishLinked))
	if put.PublishLinked == "" || untap.Object != want ||
		effectConditionMatch(t, sequence[2]).Object.Val != want {
		t.Fatal("modal condition and consequence do not share the actual returned incarnation")
	}
	assertCardPaths(t, card, "SpellAbility.Val.Modes[0].Sequence")
}
