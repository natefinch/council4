package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func exileCreatureFromGraveyardInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.ExileFromGraveyard{
			Player:    game.ControllerReference(),
			Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
			Amount:    game.Fixed(1),
		},
	}
}

// TestExileFromGraveyardExilesChosenCreature verifies the controller's chosen
// matching creature card moves from graveyard to exile while a non-matching card
// stays in the graveyard.
func TestExileFromGraveyardExilesChosenCreature(t *testing.T) {
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
	engine.resolveInstructionWithChoices(g, obj, exileCreatureFromGraveyardInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(creature) {
		t.Fatal("chosen creature was not exiled")
	}
	if g.Players[game.Player1].Graveyard.Contains(creature) {
		t.Fatal("chosen creature still in graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(instant) {
		t.Fatal("non-matching instant was removed from graveyard")
	}
}

// TestExileFromGraveyardWithNoMatchingCardDoesNothing verifies that with no
// matching card in the graveyard, the effect leaves the graveyard intact and
// exiles nothing.
func TestExileFromGraveyardWithNoMatchingCardDoesNothing(t *testing.T) {
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
	engine.resolveInstructionWithChoices(g, obj, exileCreatureFromGraveyardInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(instant) {
		t.Fatal("non-matching instant was removed from graveyard")
	}
	if g.Players[game.Player1].Exile.Size() != 0 {
		t.Fatalf("exile size = %d, want 0", g.Players[game.Player1].Exile.Size())
	}
}

// TestExileFromGraveyardExilesMultipleChosenCards verifies an Amount greater than
// one exiles that many matching cards from the controller's graveyard.
func TestExileFromGraveyardExilesMultipleChosenCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	bear := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	ox := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Ox",
		Types: []types.Card{types.Creature},
	}})

	instruction := &game.Instruction{
		Primitive: game.ExileFromGraveyard{
			Player:    game.ControllerReference(),
			Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
			Amount:    game.Fixed(2),
		},
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, instruction, agents, &TurnLog{})

	for _, cardID := range []id.ID{bear, ox} {
		if !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatalf("card %v was not exiled", cardID)
		}
		if g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("card %v still in graveyard", cardID)
		}
	}
}
