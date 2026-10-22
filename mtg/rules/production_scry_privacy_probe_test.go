package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestProductionScryPrivateLogBoundary(t *testing.T) {
	for _, amount := range []int{1, 3} {
		t.Run(string(rune('0'+amount)), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			for range amount {
				addCardToLibrary(g, game.Player1, vanillaCreature("Private scry identity", 1, 1))
			}
			addEffectSpellToStack(g, game.Player1, game.Scry{
				Amount: game.Fixed(amount), Player: game.ControllerReference(),
			}, nil)
			agent := &recordingChoiceAgent{answer: []int{0}}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &log)
			if len(agent.requests) == 0 || len(log.Choices) == 0 {
				t.Fatal("scry choice did not execute")
			}
			if amount == 1 && !agent.requests[0].Subject.Exists ||
				amount > 1 && !agent.requests[0].Options[0].Card.Exists {
				t.Fatal("deciding agent did not receive the private card identity")
			}
			for _, decision := range log.Choices {
				if decision.Request.Subject.Exists {
					t.Error("private scry subject leaked into returned log")
				}
				for _, option := range decision.Request.Options {
					if option.Card.Exists || option.Label == "Private scry identity" {
						t.Error("private scry/order option leaked into returned log")
					}
				}
			}
		})
	}
}
