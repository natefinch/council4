package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerContextualObjectCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		body  string
		index int
		slot  int
		kind  game.ObjectReferenceKind
	}{
		{"adjacent", "Exile target card from a graveyard. If it was a creature card, you gain 3 life.", 1, 0, game.ObjectReferenceTargetCard},
		{"nonadjacent", "Exile target card from a graveyard. You gain 1 life. If it was a creature card, draw a card.", 2, 0, game.ObjectReferenceTargetCard},
		{"second occurrence", "Tap target creature. Exile target card from a graveyard. If it was a land card, draw a card.", 2, 1, game.ObjectReferenceTargetCard},
		{"destroyed permanent", "Destroy target creature. If that creature was a Human, draw a card.", 1, 0, game.ObjectReferenceTargetPermanent},
		{"contracted", "Exile target card from a graveyard. If it's a creature card, draw a card.", 1, 0, game.ObjectReferenceTargetCard},
		{"compound contracted", "Exile target card from a graveyard. If it's an artifact creature card, draw a card.", 1, 0, game.ObjectReferenceTargetCard},
		{"target supersedes source", "Sacrifice this creature. Exile target card from a graveyard. If it was a creature card, draw a card.", 2, 0, game.ObjectReferenceTargetCard},
		{"explicit source remains source", "Exile target card from a graveyard. If this creature has a +1/+1 counter on it, you gain 1 life.", 1, 0, game.ObjectReferenceSourcePermanent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name: "Contextual Subject", Layout: "normal", TypeLine: "Creature",
				OracleText: "When this creature enters, " + test.body,
			}
			face := lowerSingleFace(t, card)
			sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
			gate := effectConditionMatch(t, sequence[test.index])
			if gate.Object.Val.Kind() != test.kind || gate.Object.Val.TargetIndex() != test.slot {
				t.Fatalf("object = %#v, want kind %v slot %d", gate.Object.Val, test.kind, test.slot)
			}
			assertCardPaths(t, card,
				"TriggeredAbilities[0].Content.Modes[0].Sequence",
				"ObjectMatches.Exists = true",
			)
			assertCardPathsAbsent(t, card, "TriggeredAbilities[0].Trigger.InterveningIf")
		})
	}
}

func TestContextualObjectConditionEventNounRemainsEvent(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Event Land Rider", Layout: "normal", TypeLine: "Creature",
		OracleText: "Whenever a land you control enters, tap target creature an opponent controls. If that land is an Island, that creature doesn't untap during its controller's next untap step.",
	}
	face := lowerSingleFace(t, card)
	gate := effectConditionMatch(t, face.TriggeredAbilities[0].Content.Modes[0].Sequence[1])
	if gate.Object.Val.Kind() != game.ObjectReferenceEventPermanent {
		t.Fatalf("that-land gate = %#v, want event permanent", gate.Object.Val)
	}
}

func TestLowerContextualObjectConditionRejectsIncompleteOwnership(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		edit func(*compiler.CompiledCondition)
		ctx  conditionLoweringContext
	}{
		{"missing reference", func(c *compiler.CompiledCondition) { c.ObjectReference = nil }, conditionContextEffectGate},
		{"different identity", func(c *compiler.CompiledCondition) { c.SubjectRefID++ }, conditionContextEffectGate},
		{"different binding", func(c *compiler.CompiledCondition) { c.ObjectBinding = compiler.ReferenceBindingSource }, conditionContextEffectGate},
		{"missing target domain", func(c *compiler.CompiledCondition) { c.ObjectTarget = nil }, conditionContextEffectGate},
		{"negative occurrence", func(c *compiler.CompiledCondition) { c.ObjectReference.Occurrence = -1 }, conditionContextEffectGate},
		{"intervening context", func(*compiler.CompiledCondition) {}, conditionContextInterveningTrigger},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			condition := compiler.CompiledCondition{
				HasSubjectReference: true, SubjectRefID: 17,
				ObjectBinding: compiler.ReferenceBindingTarget,
				ObjectReference: &compiler.CompiledReference{
					NodeID: 17, Binding: compiler.ReferenceBindingTarget,
				},
				ObjectTarget: &compiler.CompiledTarget{
					Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
					Selector:    compiler.CompiledSelector{Kind: compiler.SelectorCard, Zone: zone.Graveyard},
				},
			}
			test.edit(&condition)
			if _, ok := lowerObjectMatchReference(condition, test.ctx); ok {
				t.Fatal("incomplete contextual ownership unexpectedly lowered")
			}
		})
	}
}

func TestContextualObjectConditionAmbiguousTargetLayouts(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"Exile two target cards from a graveyard. If it was a creature card, draw a card.",
		"Tap two target creatures. Exile target card from a graveyard. If it was a land card, draw a card.",
		"Put a +1/+1 counter on up to one target creature. Exile target card from a graveyard. If it was a creature card, draw a card.",
		"Exile target card from a graveyard and target card from another graveyard. If it was a creature card, draw a card.",
	} {
		assertCardUnsupported(t, &ScryfallCard{
			Name: "Ambiguous Subject", Layout: "normal", TypeLine: "Creature",
			OracleText: "When this creature enters, " + body,
		}, "per-effect condition unrecognized")
	}
}

func TestContextualObjectConditionOwnerPayoffRemapsTarget(t *testing.T) {
	t.Parallel()
	for _, preceding := range []string{
		"Return target card from your graveyard to your hand.",
		"Tap target creature.",
	} {
		t.Run(preceding, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name: "Card Owner Rider", Layout: "normal", TypeLine: "Creature",
				OracleText: "When this creature enters, " + preceding +
					" Exile target card from an opponent's graveyard. If it was a creature card, that player loses 1 life.",
			}
			face := lowerSingleFace(t, card)
			sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
			gate := effectConditionMatch(t, sequence[2])
			payoff, ok := sequence[2].Primitive.(game.LoseLife)
			if !ok {
				t.Fatalf("payoff = %T, want LoseLife", sequence[2].Primitive)
			}
			object, ok := payoff.Player.Object()
			if !ok || object.Kind() != game.ObjectReferenceTargetCard || object.TargetIndex() != 1 ||
				gate.Object.Val.Kind() != game.ObjectReferenceTargetCard || gate.Object.Val.TargetIndex() != 1 {
				t.Fatalf("owner/gate = %#v / %#v, want card-target slot 1", object, gate.Object.Val)
			}
		})
	}
}

func TestContextualObjectConditionGrounding(t *testing.T) {
	t.Parallel()
	assertCardPaths(t, &ScryfallCard{
		Name: "Carrion Locust", Layout: "normal", TypeLine: "Creature — Insect Horror",
		OracleText: "Flying\nWhen this creature enters, exile target card from an opponent's graveyard. If it was a creature card, that player loses 1 life.",
	}, "TriggeredAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.ObjectMatches.Val.RequiredTypes[0] = types.Creature")
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Misfortune Teller", Layout: "normal", TypeLine: "Creature — Human Warlock",
		OracleText: "Deathtouch\nWhenever this creature enters or deals combat damage to a player, exile target card from a graveyard. If it was a creature card, create a 2/2 black Rogue creature token. If it was a land card, create a Treasure token. Otherwise, you gain 3 life.",
	}, "unsupported triggered ability")
	assertCardPaths(t, &ScryfallCard{
		Name: "Misfortune Teller Body", Layout: "normal", TypeLine: "Creature — Human Warlock",
		OracleText: "When this creature enters, exile target card from a graveyard. If it was a creature card, create a 2/2 black Rogue creature token.",
	}, "TriggeredAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.ObjectMatches.Val.RequiredTypes[0] = types.Creature")
}
