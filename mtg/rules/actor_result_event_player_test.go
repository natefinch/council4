package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledEventPlayerOptionalActionResult(t *testing.T) {
	for _, verb := range []string{"discard", "mill"} {
		for _, decline := range []bool{false, true} {
			t.Run(verb+map[bool]string{false: "/accepted", true: "/declined"}[decline], func(t *testing.T) {
				def := compileCounterTaxCard(t, "Event Player Result",
					"At the beginning of each player's upkeep, that player may "+verb+" a card. You gain 1 life. If they do, you gain 2 life.", "Enchantment")
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				source := addCombatPermanent(g, game.Player3, def)
				source.Controller = game.Player1
				if verb == "discard" {
					addCardToHand(g, game.Player2, vanillaCreature("Discard", 1, 1))
				} else {
					addCardToLibrary(g, game.Player2, vanillaCreature("Mill", 1, 1))
				}
				obj := &game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, SourceID: source.ObjectID,
					SourceCardID: source.CardInstanceID, Controller: game.Player1,
					InlineTrigger: &def.TriggeredAbilities[0], HasTriggerEvent: true,
					TriggerEvent: game.Event{Kind: game.EventBeginningOfStep, Controller: game.Player4, Player: game.Player2},
				}
				g.Stack.Push(obj)
				agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
				if decline {
					agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0}}}
				}
				NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				wantLife, wantMoved := 43, 1
				if decline {
					wantLife, wantMoved = 41, 0
				}
				if g.Players[game.Player1].Life != wantLife || g.Players[game.Player2].Graveyard.Size() != wantMoved ||
					g.Players[game.Player3].Life != 40 || g.Players[game.Player4].Life != 40 {
					t.Fatal("event player action/decision borrowed another player or receipt")
				}
				assertResultKeysCleaned(t, obj, "if-you-do", "result-clause-1")
			})
		}
	}
}
