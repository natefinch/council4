package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func returnCreatureFromGraveyardInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.ReturnFromGraveyard{
			Player:    game.ControllerReference(),
			Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
			Amount:    game.Fixed(1),
		},
	}
}

func addCardToGraveyardCreature(g *game.Game, playerID game.PlayerID, def *game.CardDef) game.ObjectID {
	cardID := addCardToHand(g, playerID, def)
	g.Players[playerID].Hand.Remove(cardID)
	g.Players[playerID].Graveyard.Add(cardID)
	return cardID
}

// TestReturnFromGraveyardReturnsChosenCreature verifies the controller's chosen
// matching creature card moves from graveyard to hand while a non-matching card
// stays in the graveyard.
func TestReturnFromGraveyardReturnsChosenCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	creature := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	instant := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bolt",
		Types: []types.Card{types.Instant},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, returnCreatureFromGraveyardInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(creature) {
		t.Fatal("chosen creature was not returned to hand")
	}
	if g.Players[game.Player1].Graveyard.Contains(creature) {
		t.Fatal("chosen creature still in graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(instant) {
		t.Fatal("non-matching instant was removed from graveyard")
	}
}

// TestReturnFromGraveyardWithNoMatchingCardDoesNothing verifies that with no
// matching card in the graveyard, the effect leaves the graveyard intact and
// returns nothing.
func TestReturnFromGraveyardWithNoMatchingCardDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	instant := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bolt",
		Types: []types.Card{types.Instant},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, returnCreatureFromGraveyardInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(instant) {
		t.Fatal("non-matching instant was removed from graveyard")
	}
	if g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0", g.Players[game.Player1].Hand.Size())
	}
}
