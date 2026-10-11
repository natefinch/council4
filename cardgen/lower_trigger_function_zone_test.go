package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestStepTriggerGraveyardSourceRequiresFunctionZoneProof(t *testing.T) {
	for _, text := range []string{
		"At the beginning of your upkeep, you may return this card from your graveyard to your hand.",
		"At the beginning of your upkeep, you may return Dormant Source from your graveyard to your hand.",
		"At the beginning of your upkeep, return Dormant Source from your graveyard to your hand.",
		"At the beginning of each upkeep, you may return Dormant Source from your graveyard to your hand.",
		"At the beginning of your upkeep, you may return Dormant Source from your graveyard to your hand. You gain 1 life.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardPaths(t, &ScryfallCard{
				Name: "Dormant Source", Layout: "normal", TypeLine: "Creature",
				OracleText: text, Power: new("1"), Toughness: new("1"),
			}, "TriggeredAbilities[0].ZoneOfFunction = zone.Graveyard")
		})
	}
	t.Run("complete original owns graveyard source", func(t *testing.T) {
		assertCardPaths(t, &ScryfallCard{
			Name: "Squee, Goblin Nabob", Layout: "normal", ManaCost: "{2}{R}", Colors: []string{"R"},
			TypeLine: "Legendary Creature — Goblin", Power: new("1"), Toughness: new("1"),
			OracleText: "At the beginning of your upkeep, you may return Squee, Goblin Nabob from your graveyard to your hand.",
		}, "TriggeredAbilities[0].ZoneOfFunction = zone.Graveyard")
	})
	t.Run("default battlefield step trigger", func(t *testing.T) {
		assertCardPaths(t, &ScryfallCard{
			Name: "Live Source", Layout: "normal", TypeLine: "Creature",
			OracleText: "At the beginning of your upkeep, you gain 1 life.", Power: new("1"), Toughness: new("1"),
		}, "TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.GainLife).Amount.fixed = 1")
	})
	t.Run("self-death uses exact event card", func(t *testing.T) {
		assertCardPaths(t, &ScryfallCard{
			Name: "Departing Source", Layout: "normal", TypeLine: "Creature",
			OracleText: "When this creature dies, return it to its owner's hand.",
			Power:      new("1"), Toughness: new("1"),
		}, "TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.MoveCard).FromZone = zone.Graveyard",
			"TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.MoveCard).Card.Kind = game.CardReferenceEvent")
	})
	t.Run("activated explicit graveyard source", func(t *testing.T) {
		assertCardPaths(t, &ScryfallCard{
			Name: "Activated Source", Layout: "normal", TypeLine: "Creature",
			OracleText: "{1}: Return this card from your graveyard to your hand.",
			Power:      new("1"), Toughness: new("1"),
		}, "ActivatedAbilities[0].ZoneOfFunction = zone.Graveyard",
			"ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.MoveCard).Card.Kind = game.CardReferenceSource")
	})
}

func TestStepTriggerGraveyardTargetDoesNotRequireSourceFunctionZone(t *testing.T) {
	def := lowerSingleFace(t, &ScryfallCard{
		Name: "Live Retriever", Layout: "normal", TypeLine: "Creature",
		OracleText: "At the beginning of your upkeep, return target creature card from your graveyard to your hand.",
		Power:      new("1"), Toughness: new("1"),
	})
	instruction := def.TriggeredAbilities[0].Content.Modes[0].Sequence[0]
	move, ok := instruction.Primitive.(game.MoveCard)
	if !ok || move.Card.Kind != game.CardReferenceTarget || move.FromZone != zone.Graveyard ||
		move.Destination != zone.Hand {
		t.Fatalf("targeted graveyard move lost its distinct subject: %#v", instruction)
	}
}

func TestRecurringGraveyardTriggerNearMissesFailClosed(t *testing.T) {
	for _, text := range []string{
		"At the beginning of your draw step, return this card from your graveyard to your hand.",
		"At the beginning of your upkeep, return this creature from your graveyard to your hand.",
		"At the beginning of your upkeep, return this card from an opponent's graveyard to your hand.",
		"At the beginning of your upkeep, return this card and another card from your graveyard to your hand.",
		"At the beginning of your upkeep, return Dormant Source and another card from your graveyard to your hand.",
		"At the beginning of your upkeep, return Foreign Source from your graveyard to your hand.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Dormant Source", Layout: "normal", TypeLine: "Creature",
				OracleText: text, Power: new("1"), Toughness: new("1"),
			})
		})
	}
}

func TestNonStepTriggerDoesNotGainGraveyardFunctionZone(t *testing.T) {
	def := lowerSingleFace(t, &ScryfallCard{
		Name: "Dormant Source", Layout: "normal", TypeLine: "Creature",
		OracleText: "Whenever you gain life, return this card from your graveyard to your hand.",
		Power:      new("1"), Toughness: new("1"),
	})
	if def.TriggeredAbilities[0].ZoneOfFunction != zone.None {
		t.Fatal("an unrelated event acquired new graveyard discovery")
	}
}
