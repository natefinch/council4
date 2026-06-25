package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerSpellCostModifierOpponentsIncrease proves the resolving,
// duration-bounded tax "Spells your opponents cast cost {2} more to cast until
// your next turn." lowers to a one-shot ApplyRule carrying a
// RuleEffectCostModifier that raises the generic cost of opponents' spells until
// the controller's next turn (issue #1500).
func TestLowerSpellCostModifierOpponentsIncrease(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Opponent Tax",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{W}",
		OracleText: "Spells your opponents cast cost {2} more to cast until your next turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	apply := requireApplyRule(t, face.SpellAbility.Val.Modes[0])
	if apply.Duration != game.DurationUntilYourNextTurn {
		t.Fatalf("duration = %v, want DurationUntilYourNextTurn", apply.Duration)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectCostModifier {
		t.Fatalf("kind = %v, want RuleEffectCostModifier", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerOpponent {
		t.Fatalf("affected player = %v, want PlayerOpponent", effect.AffectedPlayer)
	}
	if effect.CostModifier.Kind != game.CostModifierSpell {
		t.Fatalf("modifier kind = %v, want CostModifierSpell", effect.CostModifier.Kind)
	}
	if effect.CostModifier.GenericIncrease != 2 {
		t.Fatalf("generic increase = %d, want 2", effect.CostModifier.GenericIncrease)
	}
	if effect.CostModifier.GenericReduction != 0 {
		t.Fatalf("generic reduction = %d, want 0", effect.CostModifier.GenericReduction)
	}
	if !effect.CostModifier.CardSelection.Empty() {
		t.Fatal("unexpected card-type filter on unfiltered tax")
	}
}

// TestLowerSpellCostModifierControllerReductionFiltered proves Armor Wars
// chapter II ("Artifact spells you cast this turn cost {1} less to cast.") lowers
// to a controller-scoped, this-turn cost reduction restricted to artifact
// spells.
func TestLowerSpellCostModifierControllerReductionFiltered(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Armor Discount",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{W}",
		OracleText: "Artifact spells you cast this turn cost {1} less to cast.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	apply := requireApplyRule(t, face.SpellAbility.Val.Modes[0])
	if apply.Duration != game.DurationThisTurn {
		t.Fatalf("duration = %v, want DurationThisTurn", apply.Duration)
	}
	effect := apply.RuleEffects[0]
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if effect.CostModifier.GenericReduction != 1 {
		t.Fatalf("generic reduction = %d, want 1", effect.CostModifier.GenericReduction)
	}
	if cardSel := effect.CostModifier.CardSelection; len(cardSel.RequiredTypes) != 1 || cardSel.RequiredTypes[0] != types.Artifact {
		t.Fatalf("card-type filter = %+v, want RequiredTypes [Artifact]", effect.CostModifier.CardSelection)
	}
}

// TestLowerSpellCostModifierNoncreatureExcluded proves the noncreature
// exclusion filter (Elspeth Conquers Death chapter II, "Noncreature spells your
// opponents cast cost {2} more to cast until your next turn.") lowers to a
// resolving, duration-bounded tax that raises the generic cost of opponents'
// spells, restricted to spells that are not creatures via the negative
// card-type filter.
func TestLowerSpellCostModifierNoncreatureExcluded(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Noncreature Tax",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{W}",
		OracleText: "Noncreature spells your opponents cast cost {2} more to cast until your next turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	apply := requireApplyRule(t, face.SpellAbility.Val.Modes[0])
	if apply.Duration != game.DurationUntilYourNextTurn {
		t.Fatalf("duration = %v, want DurationUntilYourNextTurn", apply.Duration)
	}
	effect := apply.RuleEffects[0]
	if effect.AffectedPlayer != game.PlayerOpponent {
		t.Fatalf("affected player = %v, want PlayerOpponent", effect.AffectedPlayer)
	}
	if effect.CostModifier.Kind != game.CostModifierSpell {
		t.Fatalf("modifier kind = %v, want CostModifierSpell", effect.CostModifier.Kind)
	}
	if effect.CostModifier.GenericIncrease != 2 {
		t.Fatalf("generic increase = %d, want 2", effect.CostModifier.GenericIncrease)
	}
	if len(effect.CostModifier.CardSelection.RequiredTypes) != 0 {
		t.Fatal("unexpected required card-type filter on noncreature tax")
	}
	if cardSel := effect.CostModifier.CardSelection; len(cardSel.ExcludedTypes) != 1 || cardSel.ExcludedTypes[0] != types.Creature {
		t.Fatalf("excluded card-type filter = %+v, want ExcludedTypes [Creature]", effect.CostModifier.CardSelection)
	}
}

func requireApplyRule(t *testing.T, mode game.Mode) game.ApplyRule {
	t.Helper()
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one primitive", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
	}
	return apply
}
