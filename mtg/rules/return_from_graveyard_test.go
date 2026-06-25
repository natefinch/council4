package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func returnCreatureFromGraveyardInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.ReturnFromGraveyardChoice(
			game.ControllerReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
			game.Fixed(1),
			zone.None,
			false,
			opt.V[int]{},
			false,
			"",
		),
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

// TestReturnFromGraveyardReanimatesChosenCreatureToBattlefield verifies the
// chosen-card reanimation path (Destination Battlefield) puts the controller's
// chosen creature onto the battlefield under their control rather than to hand.
func TestReturnFromGraveyardReanimatesChosenCreatureToBattlefield(t *testing.T) {
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

	instruction := &game.Instruction{
		Primitive: game.ReturnFromGraveyardChoice(
			game.ControllerReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
			game.Fixed(1),
			zone.Battlefield,
			false,
			opt.V[int]{},
			false,
			"",
		),
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, instruction, agents, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(creature) {
		t.Fatal("chosen creature still in graveyard")
	}
	if g.Players[game.Player1].Hand.Contains(creature) {
		t.Fatal("chosen creature went to hand instead of battlefield")
	}
	if !onBattlefieldByCard(g, creature) {
		t.Fatal("chosen creature was not put onto the battlefield")
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

func battlefieldControllerByCard(g *game.Game, cardID id.ID) (game.PlayerID, bool) {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent.Controller, true
		}
	}
	return 0, false
}

func allGraveyardsMassReturnInstruction(controlledByOwner bool) *game.Instruction {
	return &game.Instruction{
		Primitive: game.MassReturnFromGraveyard{
			Player:            game.ControllerReference(),
			Selection:         game.Selection{RequiredTypes: []types.Card{types.Creature}},
			Destination:       zone.Battlefield,
			SourceGroup:       game.AllPlayersReference(),
			ControlledByOwner: controlledByOwner,
		},
	}
}

// TestMassReturnFromGraveyardAllGraveyardsUnderYourControl verifies the
// all-graveyards reanimation (Rise of the Dark Realms) reaches every player's
// graveyard and enters each creature under the resolving controller's control.
func TestMassReturnFromGraveyardAllGraveyardsUnderYourControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	mine := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	theirs := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Ox",
		Types: []types.Card{types.Creature},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, allGraveyardsMassReturnInstruction(false), agents, &TurnLog{})

	for _, cardID := range []id.ID{mine, theirs} {
		controller, ok := battlefieldControllerByCard(g, cardID)
		if !ok {
			t.Fatalf("card %v was not put onto the battlefield", cardID)
		}
		if controller != game.Player1 {
			t.Fatalf("card %v controller = %v, want Player1", cardID, controller)
		}
	}
}

// TestMassReturnFromGraveyardAllGraveyardsUnderOwnersControl verifies the
// owners'-control all-graveyards reanimation (Open the Vaults) enters each
// creature under its own owner's control.
func TestMassReturnFromGraveyardAllGraveyardsUnderOwnersControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	mine := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	theirs := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Ox",
		Types: []types.Card{types.Creature},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, allGraveyardsMassReturnInstruction(true), agents, &TurnLog{})

	owners := map[id.ID]game.PlayerID{mine: game.Player1, theirs: game.Player2}
	for cardID, owner := range owners {
		controller, ok := battlefieldControllerByCard(g, cardID)
		if !ok {
			t.Fatalf("card %v was not put onto the battlefield", cardID)
		}
		if controller != owner {
			t.Fatalf("card %v controller = %v, want owner %v", cardID, controller, owner)
		}
	}
}
