package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerBargainKeywordAndEnterCondition proves a Bargain permanent lowers to
// the reusable BargainStaticBody plus an enters-the-battlefield triggered
// ability gated on the linked "if it was bargained" intervening-if condition
// (Troublemaker Ouphe's shape).
func TestLowerBargainKeywordAndEnterCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Troublemaker Ouphe",
		Layout:   "normal",
		TypeLine: "Creature — Ouphe",
		OracleText: "Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)\n" +
			"When this creature enters, if it was bargained, exile target artifact or enchantment an opponent controls.",
		Power:     new("2"),
		Toughness: new("2"),
	})
	if len(face.StaticAbilities) == 0 || face.StaticAbilities[0].VarName != "game.BargainStaticBody" {
		t.Fatalf("first static VarName = %+v, want game.BargainStaticBody", face.StaticAbilities)
	}
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	if !face.TriggeredAbilities[0].Trigger.InterveningIfEventPermanentWasBargained {
		t.Fatal("enter trigger is not gated on InterveningIfEventPermanentWasBargained")
	}
}

// TestLowerBargainCostReduction proves the "this spell costs {N} less to cast if
// it's bargained" rider lowers to a self-affecting spell cost modifier whose
// reduction is conditioned on SpellWasBargained (Ice Out's shape).
func TestLowerBargainCostReduction(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Ice Out",
		Layout:   "normal",
		TypeLine: "Instant",
		OracleText: "Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)\n" +
			"This spell costs {1} less to cast if it's bargained.\n" +
			"Counter target spell.",
	})
	if len(face.StaticAbilities) == 0 || face.StaticAbilities[0].VarName != "game.BargainStaticBody" {
		t.Fatalf("first static VarName = %+v, want game.BargainStaticBody", face.StaticAbilities)
	}
	var reduction *game.CostModifier
	for i := range face.StaticAbilities {
		for _, effect := range face.StaticAbilities[i].Body.RuleEffects {
			if effect.Kind == game.RuleEffectCostModifier {
				reduction = &effect.CostModifier
			}
		}
	}
	if reduction == nil {
		t.Fatal("no cost modifier rule effect lowered from the bargained cost reduction")
	}
	if reduction.GenericReduction != 1 {
		t.Fatalf("GenericReduction = %d, want 1", reduction.GenericReduction)
	}
	if !reduction.ReductionCondition.Exists || !reduction.ReductionCondition.Val.SpellWasBargained {
		t.Fatalf("reduction condition = %+v, want SpellWasBargained", reduction.ReductionCondition)
	}
}
