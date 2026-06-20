package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestMoveCommanderPutsCommanderFromCommandZoneIntoHand proves Command Beacon's
// "Put your commander into your hand from the command zone." moves the
// controller's commander out of the command zone and into their hand, bypassing
// the commander-zone replacement (CR 903.9) that would otherwise redirect it.
func TestMoveCommanderPutsCommanderFromCommandZoneIntoHand(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{
		game.Player1: {Commander: commanderDef("Zone Commander", color.Green)},
	})
	engine := NewEngine(nil)
	commanderID := g.Players[game.Player1].CommanderInstanceID
	if commanderID == 0 || !g.Players[game.Player1].CommandZone.Contains(commanderID) {
		t.Fatalf("commander %v not in command zone at setup", commanderID)
	}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.MoveCommander{
			Player:      game.ControllerReference(),
			Destination: zone.Hand,
		},
	}}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].CommandZone.Contains(commanderID) {
		t.Fatal("commander remained in command zone (the move was redirected)")
	}
	if !g.Players[game.Player1].Hand.Contains(commanderID) {
		t.Fatal("commander was not placed into its owner's hand")
	}
}

// TestMoveCommanderLeavesOtherPlayersCommandersUntouched proves the effect only
// relocates the resolving controller's own commander.
func TestMoveCommanderLeavesOtherPlayersCommandersUntouched(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{
		game.Player1: {Commander: commanderDef("Mine", color.Green)},
		game.Player2: {Commander: commanderDef("Theirs", color.Red)},
	})
	engine := NewEngine(nil)
	mine := g.Players[game.Player1].CommanderInstanceID
	theirs := g.Players[game.Player2].CommanderInstanceID
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.MoveCommander{
			Player:      game.ControllerReference(),
			Destination: zone.Hand,
		},
	}}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(mine) {
		t.Fatal("controller's commander was not moved to hand")
	}
	if !g.Players[game.Player2].CommandZone.Contains(theirs) {
		t.Fatal("another player's commander was moved")
	}
}
