package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
)

func TestPaidCostNumericPredicatesCompose(t *testing.T) {
	for _, test := range []struct {
		name, cost, predicate, path string
	}{
		{"sacrifice toughness", "Sacrifice a creature", "the sacrificed creature had toughness 4 or greater", "Toughness"},
		{"sacrifice possessive power", "Sacrifice a creature", "the sacrificed creature's power was 4 or greater", "Power"},
		{"discard mana value", "Discard a card", "the discarded card had mana value 4 or greater", "ManaValue"},
		{"discard possessive toughness", "Discard a creature card", "the discarded card's toughness was 4 or greater", "Toughness"},
	} {
		t.Run(test.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Numeric Paid Subject", Layout: "normal", TypeLine: "Artifact",
				OracleText: test.cost + ": You gain 1 life. If " + test.predicate + ", draw a card."}
			face := lowerSingleFace(t, card)
			ability := face.ActivatedAbilities[0]
			condition := effectConditionMatch(t, ability.Content.Modes[0].Sequence[1])
			if condition.Object.Val.Kind() != game.ObjectReferencePaidCost ||
				condition.Object.Val.CostKey() != ability.AdditionalCosts[0].SubjectKey {
				t.Fatal("numeric subject did not preserve exact cost-component binding")
			}
			assertCardPaths(t, card, "ObjectMatches.Val."+test.path+".Exists = true",
				"ObjectMatches.Val."+test.path+".Val.Op = compare.GreaterOrEqual",
				"ObjectMatches.Val."+test.path+".Val.Value = 4")
			assertCardPathsAbsent(t, card, "ActivationCondition.Exists = true", "TargetCardResultKey", "ResultGate")
		})
	}
}

func TestPaidCostNumericCurrentAndRelativeRefuse(t *testing.T) {
	for _, predicate := range []string{
		"the sacrificed creature has toughness 4 or greater",
		"the sacrificed creature's toughness is 4 or greater",
		"the sacrificed creature had toughness greater than this creature's power",
		"the discarded card has mana value 4 or greater",
		"the discarded card's mana value is 4 or greater",
		"the discarded card had the same mana value as target spell",
	} {
		t.Run(predicate, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{Name: "Unavailable Numeric Subject", Layout: "normal", TypeLine: "Artifact",
				OracleText: "Sacrifice a creature, discard a card: If " + predicate + ", draw a card."})
		})
	}
}

func TestPaidCostNumericSharedInsteadEnvelope(t *testing.T) {
	card := &ScryfallCard{Name: "Numeric Replacement", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "As an additional cost to cast this spell, sacrifice a creature.\nDraw two cards. If the sacrificed creature had toughness 4 or greater, draw three cards instead."}
	face := lowerSingleFace(t, card)
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	found := false
	for _, instruction := range sequence {
		if !instruction.Condition.Exists || !instruction.Condition.Val.Condition.Exists {
			continue
		}

		condition := instruction.Condition.Val.Condition.Val
		if condition.Object.Exists && condition.Object.Val.Kind() == game.ObjectReferencePaidCost {
			found = condition.ObjectMatches.Val.Toughness.Val == (compare.Int{Op: compare.GreaterOrEqual, Value: 4})
		}
	}
	if !found {
		t.Fatal("shared Instead envelope lost the exact actual-cost numeric predicate")
	}
}

func TestPaidCostNumericTokenInsteadEnvelope(t *testing.T) {
	card := &ScryfallCard{Name: "Numeric Token Replacement", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{T}, Sacrifice a creature: Create a Food token. If the sacrificed creature had toughness 4 or greater, create two Food tokens instead."}
	assertCardPaths(t, card, "Object.Val.kind = game.ObjectReferencePaidCost",
		"ObjectMatches.Val.Toughness.Val.Value = 4", "Subtypes[0] = types.Food")
}
