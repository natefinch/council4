package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// mowuSelfCounterReplacementCardDef mirrors Mowu, Loyal Companion's self-scoped
// "that many plus one" +1/+1 counter replacement.
func mowuSelfCounterReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Mowu, Loyal Companion",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.SelfCounterPlacementReplacement(
				"If one or more +1/+1 counters would be put on Mowu, that many plus one +1/+1 counters are put on it instead.",
				0,
				1,
				counter.PlusOnePlusOne,
			),
		},
	}}
}

// TestSelfCounterReplacementOnlyAffectsItsOwnSource proves the self-scoped
// counter-amount replacement adds its extra counter only when counters are put
// on the replacement's own permanent, leaving every other creature the
// controller controls untouched ("on Mowu", not "on a creature you control").
func TestSelfCounterReplacementOnlyAffectsItsOwnSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mowu := addReplacementPermanent(t, g, game.Player1, mowuSelfCounterReplacementCardDef())
	other := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Other Dog",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, mowu, counter.PlusOnePlusOne, 2) {
		t.Fatal("addCountersToPermanent(mowu) = false, want true")
	}
	if got := mowu.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters on Mowu = %d, want 3 (2 plus one)", got)
	}
	if !addCountersToPermanent(g, other, counter.PlusOnePlusOne, 2) {
		t.Fatal("addCountersToPermanent(other) = false, want true")
	}
	if got := other.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters on other creature = %d, want 2 (not boosted)", got)
	}
}
