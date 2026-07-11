package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// burgeoningLandTrigger gives controller a permanent that models Burgeoning:
// "Whenever an opponent plays a land, you may put a land card from your hand
// onto the battlefield." The optional ability fires on the EventLandPlayed
// runtime event, scoped to land plays by an opponent.
func burgeoningLandTrigger(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addOptionalTriggeredPermanent(g, controller,
		&game.TriggerPattern{Event: game.EventLandPlayed, Player: game.TriggerPlayerOpponent},
		[]game.Instruction{{
			Primitive: game.ChooseFromZone{
				Player:      game.ControllerReference(),
				SourceZone:  zone.Hand,
				Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}},
				Quantity:    game.Fixed(1),
				Destination: game.ChooseDestination{Zone: zone.Battlefield},
				Prompt:      "Choose a card to put onto the battlefield",
			},
		}}, nil)
}

// TestPlaysLandTriggerMatchesOpponentNotSelf proves the EventLandPlayed pattern
// with TriggerPlayerOpponent matches a land played by an opponent but not one
// played by the trigger's own controller.
func TestPlaysLandTriggerMatchesOpponentNotSelf(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := burgeoningLandTrigger(g, game.Player1)
	pattern := &game.TriggerPattern{Event: game.EventLandPlayed, Player: game.TriggerPlayerOpponent}

	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:       game.EventLandPlayed,
		Controller: game.Player2,
		Player:     game.Player2,
	}) {
		t.Fatal("trigger did not match a land played by an opponent")
	}
	if triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:       game.EventLandPlayed,
		Controller: game.Player1,
		Player:     game.Player1,
	}) {
		t.Fatal("trigger matched a land played by its own controller")
	}
}

// TestBurgeoningPutsLandWhenOpponentPlaysLand exercises the real land-play path:
// an opponent plays a land, the EventLandPlayed runtime event fires Burgeoning's
// optional trigger, and accepting it moves a land from the controller's hand
// onto the battlefield.
func TestBurgeoningPutsLandWhenOpponentPlaysLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	burgeoningLandTrigger(g, game.Player1)

	landInHand := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})
	opponentLand := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Island", Types: []types.Card{types.Land}}})

	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {0}}}}
	if !engine.applyActionWithChoices(g, game.Player2, action.PlayLand(opponentLand), agents, &TurnLog{}) {
		t.Fatal("opponent land play was rejected")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("land-played trigger was not put on the stack for an opponent's land play")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(landInHand) {
		t.Fatal("land was not moved out of the controller's hand")
	}
	if permanentForCard(g, landInHand) == nil {
		t.Fatal("land from hand did not enter the battlefield")
	}
}

// TestBurgeoningDoesNotTriggerOnControllersOwnLandPlay proves Burgeoning does
// not fire when its own controller plays a land, because it triggers only on an
// opponent's land play.
func TestBurgeoningDoesNotTriggerOnControllersOwnLandPlay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	burgeoningLandTrigger(g, game.Player1)

	landInHand := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})
	ownLand := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Island", Types: []types.Card{types.Land}}})

	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {0}}}}
	if !engine.applyActionWithChoices(g, game.Player1, action.PlayLand(ownLand), agents, &TurnLog{}) {
		t.Fatal("controller land play was rejected")
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("land-played trigger fired on the controller's own land play")
	}
	if !g.Players[game.Player1].Hand.Contains(landInHand) {
		t.Fatal("controller's spare land left hand despite no trigger firing")
	}
}
