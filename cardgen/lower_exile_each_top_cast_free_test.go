package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerEtaliPrimalStormExilesEachTopAndCastsFree proves the attack-trigger
// body "exile the top card of each player's library, then you may cast any
// number of spells from among those cards without paying their mana costs."
// lowers to one ExileTopEachLibraryCastFree primitive with a per-library exile
// count of one.
func TestLowerEtaliPrimalStormExilesEachTopAndCastsFree(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Etali, Primal Storm",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Elder Dinosaur",
		OracleText: "Whenever Etali attacks, exile the top card of each player's library, then you may cast any number of spells from among those cards without paying their mana costs.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	content := face.TriggeredAbilities[0].Content
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %#v, want one instruction", content)
	}
	prim, ok := content.Modes[0].Sequence[0].Primitive.(game.ExileTopEachLibraryCastFree)
	if !ok {
		t.Fatalf("primitive = %T, want game.ExileTopEachLibraryCastFree", content.Modes[0].Sequence[0].Primitive)
	}
	if prim.Amount.IsDynamic() || prim.Amount.Value() != 1 {
		t.Fatalf("amount = %#v, want fixed one", prim.Amount)
	}
}

// TestExileEachTopCastFreeFailsClosedForBoundedCast proves the recognizer fails
// closed when the free cast is restricted by mana value ("cast any number of
// spells with mana value 3 or less"), a bound the primitive cannot express, so
// the card is reported unsupported rather than casting cards the text forbids.
func TestExileEachTopCastFreeFailsClosedForBoundedCast(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Bounded Etali",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Elder Dinosaur",
		OracleText: "Whenever this creature attacks, exile the top card of each player's library, then you may cast any number of spells with mana value 3 or less from among those cards without paying their mana costs.",
	})
}

// TestExileEachTopCastFreeFailsClosedForSingularCast proves the recognizer fails
// closed for the singular "cast a spell" form, which lacks the "any number of"
// count the ExileTopEachLibraryCastFree primitive models.
func TestExileEachTopCastFreeFailsClosedForSingularCast(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Singular Etali",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Elder Dinosaur",
		OracleText: "Whenever this creature attacks, exile the top card of each player's library, then you may cast a spell from among those cards without paying its mana cost.",
	})
}
