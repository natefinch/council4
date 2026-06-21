package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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

func massReturnInstruction(destination zone.Type) *game.Instruction {
	return &game.Instruction{
		Primitive: game.MassReturnFromGraveyard{
			Player:      game.ControllerReference(),
			Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
			Destination: destination,
		},
	}
}

// TestMassReturnFromGraveyardToBattlefield verifies every matching creature card
// in the controller's graveyard enters the battlefield at once while a
// non-matching card stays in the graveyard.
func TestMassReturnFromGraveyardToBattlefield(t *testing.T) {
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
	bolt := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bolt",
		Types: []types.Card{types.Instant},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, massReturnInstruction(zone.Battlefield), agents, &TurnLog{})

	for _, cardID := range []id.ID{bear, ox} {
		if !onBattlefieldByCard(g, cardID) {
			t.Fatalf("card %v was not put onto the battlefield", cardID)
		}
		if g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("card %v still in graveyard", cardID)
		}
	}
	if !g.Players[game.Player1].Graveyard.Contains(bolt) {
		t.Fatal("non-matching instant left the graveyard")
	}
}

// TestMassReturnFromGraveyardToHand verifies every matching creature card moves
// to its owner's hand while a non-matching card stays in the graveyard.
func TestMassReturnFromGraveyardToHand(t *testing.T) {
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
	bolt := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bolt",
		Types: []types.Card{types.Instant},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, massReturnInstruction(zone.Hand), agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(bear) {
		t.Fatal("matching creature was not returned to hand")
	}
	if g.Players[game.Player1].Graveyard.Contains(bear) {
		t.Fatal("matching creature still in graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(bolt) {
		t.Fatal("non-matching instant left the graveyard")
	}
}

func onBattlefieldByCard(g *game.Game, cardID id.ID) bool {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return true
		}
	}
	return false
}
