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

// TestLowerGroupMustAttackUntilYourNextTurn proves that the duration-scoped
// chapter effect "Until your next turn, creatures your opponents control attack
// each combat if able." (The Akroan War chapter II) lowers to an ApplyRule
// carrying RuleEffectMustAttack scoped to the opponents' creatures with an
// until-your-next-turn duration.
func TestLowerGroupMustAttackUntilYourNextTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Akroan",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Draw a card.\n" +
			"II — Draw a card.\n" +
			"III — Until your next turn, creatures your opponents control attack each combat if able.",
	})
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("chapter abilities = %d, want 3", len(face.ChapterAbilities))
	}
	mode := face.ChapterAbilities[2].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("chapter III sequence len = %d, want 1", len(mode.Sequence))
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	if apply.Duration != game.DurationUntilYourNextTurn {
		t.Fatalf("duration = %v, want DurationUntilYourNextTurn", apply.Duration)
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

// staticMustAttackEffect extracts the single RuleEffectMustAttack carried by a
// face's single static ability, failing the test if the shape differs.
func staticMustAttackEffect(t *testing.T, face loweredFaceAbilities) game.RuleEffect {
	t.Helper()
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 1 {
		t.Fatalf("rule effects = %#v, want one", effects)
	}
	if effects[0].Kind != game.RuleEffectMustAttack {
		t.Fatalf("kind = %v, want RuleEffectMustAttack", effects[0].Kind)
	}
	return effects[0]
}

// TestLowerStaticMustAttackOpponents proves that the continuous static
// "Creatures your opponents control attack each combat if able." (Angler
// Turtle) lowers to a standing RuleEffectMustAttack scoped, via the affected
// Selection, to the controller's opponents' creatures.
func TestLowerStaticMustAttackOpponents(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Angler Turtle",
		Layout:     "normal",
		TypeLine:   "Creature — Turtle",
		ManaCost:   "{4}{U}",
		OracleText: "Creatures your opponents control attack each combat if able.",
	})
	effect := staticMustAttackEffect(t, face)
	if effect.AffectedSelection.Controller != game.ControllerOpponent {
		t.Fatalf("affected selection controller = %v, want ControllerOpponent", effect.AffectedSelection.Controller)
	}
	if len(effect.PermanentTypes) != 1 || effect.PermanentTypes[0] != types.Creature {
		t.Fatalf("permanent types = %v, want [Creature]", effect.PermanentTypes)
	}
}

// TestLowerStaticMustAttackControlled proves that "Creatures you control attack
// each combat if able." (Thantis the Warweaver) scopes the standing
// forced-attack rule to the controller's own creatures.
func TestLowerStaticMustAttackControlled(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Thantis",
		Layout:     "normal",
		TypeLine:   "Creature — Insect",
		ManaCost:   "{3}{B}{G}",
		OracleText: "Creatures you control attack each combat if able.",
	})
	effect := staticMustAttackEffect(t, face)
	if effect.AffectedController != game.ControllerYou {
		t.Fatalf("affected controller = %v, want ControllerYou", effect.AffectedController)
	}
}

// TestLowerStaticMustAttackAll proves that the all-creatures static form
// (Avatar of Slaughter's "All creatures attack each combat if able.") scopes
// the standing forced-attack rule to every creature regardless of controller.
func TestLowerStaticMustAttackAll(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Avatar of Slaughter",
		Layout:     "normal",
		TypeLine:   "Creature — Avatar",
		ManaCost:   "{4}{R}{R}",
		OracleText: "All creatures attack each combat if able.",
	})
	effect := staticMustAttackEffect(t, face)
	if effect.AffectedController != game.ControllerAny {
		t.Fatalf("affected controller = %v, want ControllerAny", effect.AffectedController)
	}
	if len(effect.PermanentTypes) != 1 || effect.PermanentTypes[0] != types.Creature {
		t.Fatalf("permanent types = %v, want [Creature]", effect.PermanentTypes)
	}
}
