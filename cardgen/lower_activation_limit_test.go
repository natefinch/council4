package cardgen

import (
	"testing"
)

// TestLowerActivateNoMoreThanTwicePerTurn verifies "Activate no more than twice
// each turn." (Pit Imp) lowers to an activated ability capped at two activations
// per turn, leaving the pump body otherwise unchanged.
func TestLowerActivateNoMoreThanTwicePerTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Pit Imp",
		Layout:     "normal",
		TypeLine:   "Creature — Imp",
		ManaCost:   "{B}",
		OracleText: "Flying\n{B}: This creature gets +1/+0 until end of turn. Activate no more than twice each turn.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	if got := face.ActivatedAbilities[0].MaxActivationsPerTurn; got != 2 {
		t.Fatalf("MaxActivationsPerTurn = %d, want 2", got)
	}
}

// TestLowerActivateNoMoreThanCapOnManaAbilityFailsClosed verifies a per-turn
// activation cap on a mana ability ("{1}: Add {B} or {R}. Activate no more than
// three times each turn.", Manaforge Cinder) fails closed: the cap is not
// enforced on the mana-payment path, so lowering must not silently ignore it.
func TestLowerActivateNoMoreThanCapOnManaAbilityFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Manaforge Cinder",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		ManaCost:   "{B}",
		OracleText: "{1}: Add {B} or {R}. Activate no more than three times each turn.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
}
