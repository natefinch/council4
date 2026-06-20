package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestRandomDiscardRemovesCardWithoutChoice(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(1), AtRandom: true}},
	})

	agent := &orderedHandChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)

	if hand := g.Players[game.Player1].Hand.All(); len(hand) != 1 {
		t.Fatalf("hand size = %d, want 1 after random discard", len(hand))
	}
	if len(agent.requests) != 0 {
		t.Fatalf("choice requests = %#v, want none for an at-random discard", agent.requests)
	}
}
