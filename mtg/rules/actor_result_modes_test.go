package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledActorResultSelectedModesIsolateActions(t *testing.T) {
	body := "Target player gains 1 life. That player discards a card. You gain 1 life. If that player does, you gain 2 life."
	def := compileCounterTaxCard(t, "Actor Mode Isolation", "Choose two —\n• "+body+"\n• "+body, "Instant")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player2, vanillaCreature("Only First Action", 1, 1))
	source := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].Hand.Remove(source)
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: source, Controller: game.Player1,
		ChosenModes: []int{0, 1}, Targets: []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)},
		TargetCounts: []int{1, 1},
	}
	g.Stack.Push(obj)
	NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Life != 44 || g.Players[game.Player2].Life != 41 ||
		g.Players[game.Player3].Life != 41 || g.Players[game.Player2].Graveyard.Size() != 1 {
		t.Fatal("second mode borrowed the first actor's actual success")
	}
	assertResultKeysCleaned(t, obj, "if-you-do", "result-clause-2")
}
