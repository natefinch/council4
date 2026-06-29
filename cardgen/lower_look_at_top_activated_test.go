package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lanternOfRevealingOracleText is an activated ability whose resolving body is
// the conditional look-at-top battlefield sequence with the optional bottom-of-
// library fallback, proving the activated form lowers to the same instruction
// template as the triggered form and that EntryTapped and the optional bottom
// fallback are carried independently.
const lanternOfRevealingOracleText = "{T}: Add one mana of any color.\n" +
	"{4}, {T}: Look at the top card of your library. If it's a land card, " +
	"you may put it onto the battlefield tapped. If you don't put the card " +
	"onto the battlefield, you may put it on the bottom of your library."

// nyamiStylePermanentOracleText is the conditional look-at-top battlefield
// sequence gated on a "permanent card" condition, proving the recognizer expands
// "permanent" to every permanent card type.
const nyamiStylePermanentOracleText = "Whenever this creature attacks, look at the top card of your library. " +
	"If it's a permanent card, you may put it onto the battlefield. " +
	"If you don't put it onto the battlefield, put it into your hand."

// TestLowerActivatedConditionalLookAtTopBattlefieldBottom proves an activated
// body lowers the look-at-top battlefield sequence into a tapped battlefield put
// whose declined card may go to the bottom of the library.
func TestLowerActivatedConditionalLookAtTopBattlefieldBottom(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Lantern of Revealing",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{3}",
		OracleText: lanternOfRevealingOracleText,
	})

	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want the look-at-top sequence", len(face.ActivatedAbilities))
	}
	sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want look/conditional-destination", len(sequence))
	}
	if _, ok := sequence[0].Primitive.(game.LookAtLibraryTop); !ok {
		t.Fatalf("sequence[0] = %#v, want LookAtLibraryTop", sequence[0].Primitive)
	}
	place, ok := sequence[1].Primitive.(game.ConditionalDestinationPlace)
	if !ok {
		t.Fatalf("sequence[1] = %#v, want ConditionalDestinationPlace", sequence[1].Primitive)
	}
	if place.Then != zone.None || !place.EntryTapped {
		t.Fatalf("conditional destination = %#v, want a tapped battlefield put", place)
	}
	if place.Else != zone.Library || !place.ElseBottom || !place.ElseOptional {
		t.Fatalf("conditional destination else = %#v, want an optional bottom fallback", place)
	}
}

// TestLowerConditionalLookAtTopBattlefieldPermanent proves a "permanent card"
// condition expands to every permanent card type.
func TestLowerConditionalLookAtTopBattlefieldPermanent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Mother Ship",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		ManaCost:   "{4}",
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: nyamiStylePermanentOracleText,
	})

	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	place, ok := sequence[1].Primitive.(game.ConditionalDestinationPlace)
	if !ok {
		t.Fatalf("sequence[1] = %#v, want ConditionalDestinationPlace", sequence[1].Primitive)
	}
	want := []types.Card{
		types.Artifact, types.Battle, types.Creature,
		types.Enchantment, types.Land, types.Planeswalker,
	}
	got := place.CardCondition.Val.Selection.RequiredTypesAny
	if len(got) != len(want) {
		t.Fatalf("required types = %#v, want every permanent type", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("required types = %#v, want %#v", got, want)
		}
	}
	if place.Else != zone.Hand {
		t.Fatalf("conditional destination else = %v, want a hand fallback", place.Else)
	}
}
