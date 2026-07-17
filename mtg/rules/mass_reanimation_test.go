package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// decliningAgent declines every optional choice (returning no selection when the
// request permits zero picks) and passes on priority, letting a test exercise the
// "up to one" path where the chooser takes nothing.
type decliningAgent struct{}

func (decliningAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (decliningAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.MinChoices == 0 {
		return []int{}
	}
	return []int{0}
}

// massReanimationSourceDef is a minimal stand-in for a mass reanimation spell
// source permanent used to drive the choose/reanimate primitives at resolution.
func massReanimationSourceDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Mass Reanimation Source",
		Types: []types.Card{types.Sorcery},
	}}
}

// TestChooseAndReanimateFromEachGraveyard verifies the mass reanimation base: the
// controller picks one matching card in every player's graveyard (one candidate
// per player forces the pick), the linked set spans one card per player, and the
// paired ReanimateLinkedCards puts exactly those cards onto the battlefield at
// once under the controller's control while leaving non-matching graveyard cards
// untouched and clearing the link.
func TestChooseAndReanimateFromEachGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, massReanimationSourceDef())
	obj := linkedSourceObject(source)

	mine := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "My Bear",
		Types: []types.Card{types.Creature},
	}})
	theirs := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Their Bear",
		Types: []types.Card{types.Creature},
	}})
	myInstant := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "My Bolt",
		Types: []types.Card{types.Instant},
	}})

	agents := [game.NumPlayers]PlayerAgent{}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ChooseCardFromEachGraveyard{
		Chooser:   game.ControllerReference(),
		Players:   game.AllPlayersReference(),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		LinkedKey: game.LinkedKey("chosen"),
	}}, agents, &TurnLog{})

	key := linkedObjectSourceKey(g, obj, "chosen")
	if got := len(linkedObjects(g, key)); got != 2 {
		t.Fatalf("linked chosen cards = %d, want 2 (one per player)", got)
	}

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ReanimateLinkedCards{
		Controller: game.ControllerReference(),
		LinkedKey:  game.LinkedKey("chosen"),
	}}, agents, &TurnLog{})

	minePerm := permanentByCardID(g, mine)
	if minePerm == nil {
		t.Fatal("Player1's graveyard creature did not enter the battlefield")
	}
	if minePerm.Controller != game.Player1 {
		t.Errorf("reanimated own creature controller = %v, want Player1", minePerm.Controller)
	}
	theirsPerm := permanentByCardID(g, theirs)
	if theirsPerm == nil {
		t.Fatal("Player2's graveyard creature did not enter the battlefield")
	}
	if theirsPerm.Controller != game.Player1 {
		t.Errorf("reanimated opponent creature controller = %v, want Player1 (under your control)", theirsPerm.Controller)
	}
	if theirsPerm.CardInstanceID != theirs {
		t.Error("wrong card reanimated for Player2")
	}
	if g.Players[game.Player1].Graveyard.Contains(mine) {
		t.Error("reanimated own creature still in graveyard")
	}
	if g.Players[game.Player2].Graveyard.Contains(theirs) {
		t.Error("reanimated opponent creature still in graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(myInstant) {
		t.Error("non-creature graveyard card was disturbed")
	}
	if got := len(linkedObjects(g, key)); got != 0 {
		t.Errorf("linked chosen cards after reanimation = %d, want 0 (consumed)", got)
	}
}

// TestChooseFromEachGraveyardOptionalDeclines verifies the optional "up to one"
// form lets the chooser decline: with Optional set and a declining agent, no card
// is linked and the paired reanimation puts nothing onto the battlefield.
func TestChooseFromEachGraveyardOptionalDeclines(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, massReanimationSourceDef())
	obj := linkedSourceObject(source)

	bear := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Optional Bear",
		Types: []types.Card{types.Creature},
	}})

	// A declining agent chooses nothing whenever declining is allowed.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: decliningAgent{}}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ChooseCardFromEachGraveyard{
		Chooser:   game.ControllerReference(),
		Players:   game.AllPlayersReference(),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Optional:  true,
		LinkedKey: game.LinkedKey("chosen"),
	}}, agents, &TurnLog{})

	key := linkedObjectSourceKey(g, obj, "chosen")
	if got := len(linkedObjects(g, key)); got != 0 {
		t.Fatalf("linked chosen cards = %d, want 0 (chooser declined the optional pick)", got)
	}

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ReanimateLinkedCards{
		Controller: game.ControllerReference(),
		LinkedKey:  game.LinkedKey("chosen"),
	}}, agents, &TurnLog{})

	if permanentByCardID(g, bear) != nil {
		t.Error("declined creature entered the battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(bear) {
		t.Error("declined creature left the graveyard")
	}
}

// graveyard card contributes nothing: only the player whose graveyard holds an
// eligible creature yields a linked card, and that lone card is reanimated.
func TestChooseFromEachGraveyardSkipsEmptyPools(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, massReanimationSourceDef())
	obj := linkedSourceObject(source)

	// Player1's graveyard holds only a non-creature card; Player2's holds a
	// creature. No prompt is needed because neither player offers a real choice.
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Lone Bolt",
		Types: []types.Card{types.Instant},
	}})
	theirs := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Lone Bear",
		Types: []types.Card{types.Creature},
	}})

	agents := [game.NumPlayers]PlayerAgent{}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ChooseCardFromEachGraveyard{
		Chooser:   game.ControllerReference(),
		Players:   game.AllPlayersReference(),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		LinkedKey: game.LinkedKey("chosen"),
	}}, agents, &TurnLog{})

	key := linkedObjectSourceKey(g, obj, "chosen")
	if got := len(linkedObjects(g, key)); got != 1 {
		t.Fatalf("linked chosen cards = %d, want 1 (only Player2 had a candidate)", got)
	}

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ReanimateLinkedCards{
		Controller: game.ControllerReference(),
		LinkedKey:  game.LinkedKey("chosen"),
	}}, agents, &TurnLog{})

	if permanentByCardID(g, theirs) == nil {
		t.Fatal("the only eligible graveyard creature did not enter the battlefield")
	}
}
