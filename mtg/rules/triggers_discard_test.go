package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func discardTriggerInstructions() []game.Instruction {
	return []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}
}

func TestDiscardTriggerRequiresCreatureCardType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:         game.EventCardDiscarded,
		Player:        game.TriggerPlayerYou,
		CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}, discardTriggerInstructions(), nil)

	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Wastes", Types: []types.Card{types.Land}}})
	if !discardCardFromHand(g, game.Player1, landID) {
		t.Fatal("discardCardFromHand(land) = false")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("creature-card discard trigger fired for a land card")
	}

	creatureID := addCardToHand(g, game.Player1, greenCreature())
	if !discardCardFromHand(g, game.Player1, creatureID) {
		t.Fatal("discardCardFromHand(creature) = false")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("creature-card discard trigger did not fire for a creature card")
	}
}

func TestDiscardTriggerExcludesCreatureAndLandCardTypes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:         game.EventCardDiscarded,
		Player:        game.TriggerPlayerYou,
		CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature, types.Land}},
	}, discardTriggerInstructions(), nil)

	creatureID := addCardToHand(g, game.Player1, greenCreature())
	if !discardCardFromHand(g, game.Player1, creatureID) {
		t.Fatal("discardCardFromHand(creature) = false")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("noncreature-nonland discard trigger fired for a creature card")
	}

	sorceryID := addCardToHand(g, game.Player1, greenSorcery())
	if !discardCardFromHand(g, game.Player1, sorceryID) {
		t.Fatal("discardCardFromHand(sorcery) = false")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("noncreature-nonland discard trigger did not fire for a sorcery card")
	}
}
