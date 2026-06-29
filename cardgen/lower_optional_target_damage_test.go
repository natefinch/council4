package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerOptionalTargetDamage covers the "deals N damage to up to one target"
// broadening: a single optional damage target lowers to one Damage instruction
// whose recipient is the optional target, with cardinality 0..1.
func TestLowerOptionalTargetDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spark",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Test Spark deals 3 damage to up to one target creature or planeswalker.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(mode.Targets))
	}
	if mode.Targets[0].MinTargets != 0 || mode.Targets[0].MaxTargets != 1 {
		t.Fatalf("cardinality = %d..%d, want 0..1", mode.Targets[0].MinTargets, mode.Targets[0].MaxTargets)
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
	}
	if damage.Amount.Value() != 3 {
		t.Fatalf("damage amount = %d, want 3", damage.Amount.Value())
	}
}

// TestLowerOptionalTargetDamageMandatoryUnchanged confirms the mandatory single
// target form still lowers to cardinality 1..1, so the optional broadening did
// not regress the existing "deals N damage to target creature" shape.
func TestLowerOptionalTargetDamageMandatoryUnchanged(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Jolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Test Jolt deals 3 damage to target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if mode.Targets[0].MinTargets != 1 || mode.Targets[0].MaxTargets != 1 {
		t.Fatalf("cardinality = %d..%d, want 1..1", mode.Targets[0].MinTargets, mode.Targets[0].MaxTargets)
	}
}
