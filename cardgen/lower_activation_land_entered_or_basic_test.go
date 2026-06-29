package cardgen

import "testing"

// TestLowerActivatedLandEnteredThisTurnOrControlsBasic proves the Mercadian
// Masques dual-mana land gate "Activate only if this land entered this turn or
// if you control a basic land." lowers to the disjunctive land activation
// condition. Previously its conjoined "or if you control a basic land" tail
// blocked the whole gate as an unsupported activation condition.
func TestLowerActivatedLandEnteredThisTurnOrControlsBasic(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bastion",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{T}: Add {W} or {U}. Activate only if this land entered this turn or if you control a basic land.",
	})
	ability := face.ManaAbilities[1]
	if !ability.ActivationCondition.Exists || !ability.ActivationCondition.Val.LandEnteredThisTurnOrControlsBasicLand {
		t.Fatalf("activation condition = %#v, want land-entered-or-controls-basic gate", ability.ActivationCondition)
	}
}
