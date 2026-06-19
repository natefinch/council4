package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerTapStunMultiTargetSequence verifies that the multi-target "tap then
// stun" spell ("Tap up to two target creatures. Those creatures don't untap
// during their controller's next untap step.") lowers to one Tap per target
// slot followed by one SkipNextUntap per slot, all bound to the single
// multi-target permanent spec carrying the 0..2 cardinality.
func TestLowerTapStunMultiTargetSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Frost Breath",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Tap up to two target creatures. Those creatures don't untap during their controller's next untap step.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want one multi-target spec", len(mode.Targets))
	}
	if mode.Targets[0].MinTargets != 0 || mode.Targets[0].MaxTargets != 2 {
		t.Fatalf("target cardinality = [%d,%d], want [0,2]", mode.Targets[0].MinTargets, mode.Targets[0].MaxTargets)
	}
	if len(mode.Sequence) != 4 {
		t.Fatalf("sequence = %d instructions, want 4 (two taps, two stuns)", len(mode.Sequence))
	}
	for slot := range 2 {
		tap, ok := mode.Sequence[slot].Primitive.(game.Tap)
		if !ok || tap.Object != game.TargetPermanentReference(slot) {
			t.Fatalf("instruction %d = %+v, want tap of target %d", slot, mode.Sequence[slot].Primitive, slot)
		}
	}
	for slot := range 2 {
		stun, ok := mode.Sequence[2+slot].Primitive.(game.SkipNextUntap)
		if !ok || stun.Object != game.TargetPermanentReference(slot) {
			t.Fatalf("instruction %d = %+v, want SkipNextUntap of target %d", 2+slot, mode.Sequence[2+slot].Primitive, slot)
		}
	}
}

// TestLowerTapStunMultiTargetFailsClosed verifies the multi-target tap-stun
// lowerer rejects shapes the SkipNextUntap primitive cannot model: a multi-step
// "next two untap steps" window, a mass tap conditioned on a target player, and
// an added third clause.
func TestLowerTapStunMultiTargetFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Tap up to two target creatures. Those creatures don't untap during their controller's next two untap steps.",
		"Tap all creatures target player controls. Those creatures don't untap during that player's next untap step.",
		"Tap up to two target creatures. Those creatures don't untap during their controller's next untap step. Draw a card.",
	}
	for _, text := range rejected {
		faces, _ := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Reject",
			Layout:     "normal",
			TypeLine:   "Instant",
			OracleText: text,
		})
		for _, face := range faces {
			if face.SpellAbility.Exists {
				t.Errorf("OracleText %q lowered a spell ability, want fail closed", text)
			}
		}
	}
}
