package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLookAtLibraryTopReadsTargetPlayerLibrary covers "look at the top card of
// target player's library." (Merfolk Observer, Dewdrop Spy): the controller
// peeks the TARGET player's library, and the informational look leaves the card
// on top of that library rather than moving it.
func TestLookAtLibraryTopReadsTargetPlayerLibrary(t *testing.T) {
	engine := NewEngine(nil)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: vanillaCreature("Peeked Card", 1, 1), Owner: game.Player2}
	g.Players[game.Player2].Library.Add(cardID)

	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	}
	resolveInstruction(engine, g, obj, game.LookAtLibraryTop{
		Player:        game.TargetPlayerReference(0),
		PublishLinked: game.LinkedKey("test-look"),
	}, nil)

	if !g.Players[game.Player2].Library.Contains(cardID) {
		t.Fatal("looked-at card was moved out of the target player's library")
	}
	if top, ok := g.Players[game.Player2].Library.Top(); !ok || top != cardID {
		t.Fatal("looked-at card no longer on top of the target player's library")
	}
}
