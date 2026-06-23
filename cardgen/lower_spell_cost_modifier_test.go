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
	if effect.CostModifier.MatchCardType {
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
	if !effect.CostModifier.MatchCardType || effect.CostModifier.CardType != types.Artifact {
		t.Fatalf("card-type filter = (%v, %v), want (true, Artifact)", effect.CostModifier.MatchCardType, effect.CostModifier.CardType)
	}
}

// TestLowerSpellCostModifierNoncreatureFailsClosed proves the noncreature
// exclusion filter (Elspeth Conquers Death chapter II) fails closed: the runtime
// spell cost modifier has no negative card-type filter, so the ability is not
// lowered to a spell.
func TestLowerSpellCostModifierNoncreatureFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Noncreature Tax",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{W}",
		OracleText: "Noncreature spells your opponents cast cost {2} more to cast until your next turn.",
	})
	if face.SpellAbility.Exists {
		t.Fatal("expected the noncreature exclusion filter to fail closed")
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
