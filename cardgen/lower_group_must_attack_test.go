package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerGroupMustAttackOpponents proves that "Creatures your opponents
// control attack this turn if able." lowers to a one-shot ApplyRule carrying
// RuleEffectMustAttack scoped to the controller's opponents' creatures for the
// turn (Bident of Thassa's activated ability).
func TestLowerGroupMustAttackOpponents(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Goad",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{R}",
		OracleText: "Creatures your opponents control attack this turn if able.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one primitive", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	if apply.Duration != game.DurationThisTurn {
		t.Fatalf("duration = %v, want DurationThisTurn", apply.Duration)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectMustAttack {
		t.Fatalf("kind = %v, want RuleEffectMustAttack", effect.Kind)
	}
	if effect.AffectedController != game.ControllerOpponent {
		t.Fatalf("affected controller = %v, want ControllerOpponent", effect.AffectedController)
	}
	if len(effect.PermanentTypes) != 1 || effect.PermanentTypes[0] != types.Creature {
		t.Fatalf("permanent types = %v, want [Creature]", effect.PermanentTypes)
	}
}

// TestLowerGroupMustAttackControlled proves that "Creatures you control attack
// this turn if able." scopes the forced-attack rule to the controller's own
// creatures.
func TestLowerGroupMustAttackControlled(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Rally",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{R}",
		OracleText: "Creatures you control attack this turn if able.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectMustAttack {
		t.Fatalf("kind = %v, want RuleEffectMustAttack", effect.Kind)
	}
	if effect.AffectedController != game.ControllerYou {
		t.Fatalf("affected controller = %v, want ControllerYou", effect.AffectedController)
	}
}

// TestLowerGroupMustAttackAll proves that "All creatures attack this turn if
// able." scopes the forced-attack rule to every creature regardless of
// controller.
func TestLowerGroupMustAttackAll(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Melee",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{R}",
		OracleText: "All creatures attack this turn if able.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	effect := apply.RuleEffects[0]
	if effect.AffectedController != game.ControllerAny {
		t.Fatalf("affected controller = %v, want ControllerAny", effect.AffectedController)
	}
}
