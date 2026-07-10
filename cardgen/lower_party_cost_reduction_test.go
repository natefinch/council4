package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerPartyCostReduction(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sea Gate Colossus",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Golem Warrior",
		ManaCost:   "{7}",
		OracleText: "This spell costs {1} less to cast for each creature in your party.",
	})
	modifier := sourceSpellCostReductionModifier(t, face)
	if modifier.DynamicReduction == nil ||
		modifier.DynamicReduction.Kind != game.DynamicAmountPartySize {
		t.Fatalf("modifier = %#v, want party-size dynamic reduction", modifier)
	}
}
