package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerPumpThenCantBeBlockedSequence verifies the common aggressive rider
// "Target creature gets +N/+0 until end of turn and can't be blocked this turn."
// (Distortion Strike, Taigam's Strike, Slip Through Space) lowers as an ordered
// sequence: the pump owns the creature target and the can't-be-blocked clause
// shares it, applying the RuleEffectCantBeBlocked to that same target. This
// exercises the ApplyRule target-index remapping the sequence machinery previously
// lacked.
func TestLowerPumpThenCantBeBlockedSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Slip",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +1/+0 until end of turn and can't be blocked this turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("mode targets = %d, want a single shared creature target", len(mode.Targets))
	}
	var applied *game.ApplyRule
	for i := range mode.Sequence {
		if rule, ok := mode.Sequence[i].Primitive.(game.ApplyRule); ok {
			applied = &rule
		}
	}
	if applied == nil {
		t.Fatalf("sequence has no ApplyRule instruction: %+v", mode.Sequence)
	}
	if !applied.Object.Exists || applied.Object.Val != game.TargetPermanentReference(0) {
		t.Fatalf("ApplyRule object = %+v, want the shared target 0", applied.Object)
	}
	if len(applied.RuleEffects) != 1 || applied.RuleEffects[0].Kind != game.RuleEffectCantBeBlocked {
		t.Fatalf("ApplyRule effects = %+v, want a single can't-be-blocked rule", applied.RuleEffects)
	}
}
