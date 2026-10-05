package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestPaidCostPredicatesCompose(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		text  string
		spell bool
		kind  game.PaidCostKind
		index int
	}{
		{"red sacrifice", "Sacrifice a creature: You gain 1 life. If the sacrificed creature was red, draw a card.", false, game.PaidCostSacrifice, 1},
		{"human sacrifice source counter", "Sacrifice a creature: Put a +1/+1 counter on this creature if the sacrificed creature was a Human.", false, game.PaidCostSacrifice, 0},
		{"legendary cast sacrifice", "As an additional cost to cast this spell, sacrifice an artifact.\nYou gain 1 life. If the sacrificed artifact was legendary, draw a card.", true, game.PaidCostSacrifice, 1},
		{"discard subtype", "Discard a creature card: You gain 1 life. If the discarded card was a Zombie card, draw a card.", false, game.PaidCostDiscard, 1},
		{"cast nonland discard", "As an additional cost to cast this spell, discard a card.\nDraw two cards. If the discarded card was not a land card, you gain 2 life.", true, game.PaidCostDiscard, 1},
		{"random multicolor discard", "Discard a card at random: You gain 1 life. If the discarded card was multicolored, draw a card.", false, game.PaidCostDiscard, 1},
		{"intervening unconditional rider", "Sacrifice a creature: Tap target creature. You gain 1 life. If the sacrificed creature was a Human, draw a card.", false, game.PaidCostSacrifice, 2},
		{"passive this way names cost", "Sacrifice a creature: You gain 1 life. If a Saproling was sacrificed this way, draw a card.", false, game.PaidCostSacrifice, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			typeLine := "Creature"
			if test.spell {
				typeLine = "Sorcery"
			}
			card := &ScryfallCard{Name: "Paid Subject", Layout: "normal", TypeLine: typeLine, OracleText: test.text}
			face := lowerSingleFace(t, card)
			var sequence []game.Instruction
			key := ""
			if test.spell {
				sequence = face.SpellAbility.Val.Modes[0].Sequence
				key = face.AdditionalCosts[0].SubjectKey
			} else {
				sequence = face.ActivatedAbilities[0].Content.Modes[0].Sequence
				key = face.ActivatedAbilities[0].AdditionalCosts[0].SubjectKey
			}
			gate := effectConditionMatch(t, sequence[test.index])
			ref := gate.Object.Val
			if key == "" || ref.Kind() != game.ObjectReferencePaidCost || ref.CostKey() != key || ref.CostKind() != test.kind {
				t.Fatalf("cost key %q, reference %#v", key, ref)
			}
			assertCardPaths(t, card, "SubjectKey = "+`"`+key+`"`, "ObjectMatches.Exists = true")
			assertCardPathsAbsent(t, card, "ActivationCondition.Exists = true", "PublishResult", "ResultGate")
		})
	}
}

func TestPaidCostSubjectRefusals(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Sacrifice two creatures: If the sacrificed creature was a Human, draw a card.",
		"Sacrifice a creature, Sacrifice an artifact: If the sacrificed creature was a Human, draw a card.",
		"Discard two cards: If the discarded card was a land card, draw a card.",
		"Sacrifice a creature or discard a card: If the discarded card was a land card, draw a card.",
		"Sacrifice a creature: If the sacrificed creature is a Human, draw a card.",
		"Discard a card: If the discarded card was suspected, draw a card.",
		"Sacrifice a creature: If the sacrificed creature was tapped, draw a card.",
		"Sacrifice a creature: If the sacrificed creature was an attacking creature, draw a card.",
		"Sacrifice a creature: If the sacrificed creature was a token creature, draw a card.",
		"You gain 1 life. If the discarded card was a land card, draw a card.",
		"Sacrifice a creature: If the sacrificed creature was a Human and you control an artifact, draw a card.",
		"Sacrifice a creature: Sacrifice another creature. If the sacrificed creature was red, draw a card.",
		"Discard a card: Discard another card. If the discarded card was a land card, draw a card.",
		"Discard a card: Discard a card. Repeat this process once. If the discarded card was a land card, draw a card.",
		"Sacrifice a creature: Sacrifice another creature. Repeat this process once. If the sacrificed creature was red, draw a card.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Unbound Paid Subject", Layout: "normal", TypeLine: "Creature", OracleText: text,
			})
		})
	}
}
