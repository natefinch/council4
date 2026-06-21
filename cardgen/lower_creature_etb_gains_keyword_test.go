package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerCreatureEtbGainsKeywordUntilEndOfTurn verifies that "Whenever a
// [filter] creature you control enters, it gains <keyword> until end of turn."
// (Dragon Tempest's first ability) lowers to an until-end-of-turn keyword grant
// whose object is the entering creature (the trigger-event permanent).
func TestLowerCreatureEtbGainsKeywordUntilEndOfTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Flying Haste Enchantment",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature you control with flying enters, it gains haste until end of turn.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want a single keyword grant", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok ||
		!apply.Object.Exists ||
		apply.Object.Val.Kind() != game.ObjectReferenceEventPermanent ||
		apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("apply = %#v, want event-permanent grant until end of turn", mode.Sequence[0].Primitive)
	}
	if len(apply.ContinuousEffects) != 1 ||
		len(apply.ContinuousEffects[0].AddKeywords) != 1 ||
		apply.ContinuousEffects[0].AddKeywords[0] != game.Haste {
		t.Fatalf("continuous effect = %#v, want haste grant", apply.ContinuousEffects)
	}
}
