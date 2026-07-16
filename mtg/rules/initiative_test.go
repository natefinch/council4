package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestTakeInitiativeSetsDesignationAndVenturesUndercity(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 5)
	agents := [game.NumPlayers]PlayerAgent{}

	if !setInitiative(g, game.Player1) {
		t.Fatal("setInitiative failed")
	}
	if !g.Players[game.Player1].HasInitiative {
		t.Fatal("player does not have the initiative")
	}
	if countEvents(g, game.EventTookInitiative) != 1 {
		t.Fatalf("took-initiative events = %d, want 1", countEvents(g, game.EventTookInitiative))
	}
	// The queued venture into Undercity resolves at the next trigger pass.
	drainDungeonStack(engine, g, agents)
	state := g.Players[game.Player1].Dungeon
	if !state.Exists || state.Val.Dungeon != game.DungeonUndercity {
		t.Fatalf("dungeon = %+v, want in Undercity after taking the initiative", state.Val)
	}
}

func TestTakeInitiativeClearsPriorHolder(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 3)
	stockLibrary(g, game.Player2, 3)
	agents := [game.NumPlayers]PlayerAgent{}

	setInitiative(g, game.Player1)
	drainDungeonStack(engine, g, agents)
	setInitiative(g, game.Player2)
	drainDungeonStack(engine, g, agents)

	if g.Players[game.Player1].HasInitiative {
		t.Fatal("prior holder still has the initiative")
	}
	if !g.Players[game.Player2].HasInitiative {
		t.Fatal("new holder does not have the initiative")
	}
}

func TestTakeInitiativeEvenIfAlreadyHolderVenturesAgain(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 8)
	agents := [game.NumPlayers]PlayerAgent{}

	setInitiative(g, game.Player1)
	drainDungeonStack(engine, g, agents)
	roomAfterFirst := g.Players[game.Player1].Dungeon.Val.Room

	// Taking the initiative again while already holding it ventures again,
	// advancing the current Undercity dungeon.
	setInitiative(g, game.Player1)
	drainDungeonStack(engine, g, agents)

	if countEvents(g, game.EventTookInitiative) != 2 {
		t.Fatalf("took-initiative events = %d, want 2", countEvents(g, game.EventTookInitiative))
	}
	if g.Players[game.Player1].Dungeon.Val.Room == roomAfterFirst {
		t.Fatal("taking the initiative while already holding it did not advance the dungeon")
	}
}

func TestCombatDamageTransfersInitiative(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	g.Players[game.Player1].HasInitiative = true

	takeInitiativeByCombatDamage(g, game.Player2, game.Player1)

	if g.Players[game.Player1].HasInitiative {
		t.Fatal("damaged holder still has the initiative")
	}
	if !g.Players[game.Player2].HasInitiative {
		t.Fatal("attacking player did not take the initiative")
	}
	if len(g.PendingInitiativeVentures) != 1 || g.PendingInitiativeVentures[0] != game.Player2 {
		t.Fatalf("pending ventures = %v, want [Player2]", g.PendingInitiativeVentures)
	}
}

func TestCombatDamageToNonHolderDoesNotTransfer(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	g.Players[game.Player3].HasInitiative = true

	// Player2 damages Player1, who does not hold the initiative.
	takeInitiativeByCombatDamage(g, game.Player2, game.Player1)

	if !g.Players[game.Player3].HasInitiative {
		t.Fatal("initiative moved off the holder on damage to a non-holder")
	}
	if g.Players[game.Player2].HasInitiative {
		t.Fatal("a player took the initiative from damage to a non-holder")
	}
}

func TestSelfCombatDamageDoesNotTransferInitiative(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	g.Players[game.Player1].HasInitiative = true

	// A creature dealing combat damage to its own controller leaves it unchanged.
	takeInitiativeByCombatDamage(g, game.Player1, game.Player1)

	if !g.Players[game.Player1].HasInitiative {
		t.Fatal("self-damage removed the initiative")
	}
}

func TestInitiativeUpkeepVenture(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	stockLibrary(g, game.Player1, 5)
	agents := [game.NumPlayers]PlayerAgent{}
	g.Players[game.Player1].HasInitiative = true

	// Model the beginning of the initiative-holder's upkeep.
	queueInitiativeVenture(g, game.Player1)
	drainDungeonStack(engine, g, agents)

	state := g.Players[game.Player1].Dungeon
	if !state.Exists || state.Val.Dungeon != game.DungeonUndercity {
		t.Fatalf("dungeon = %+v, want in Undercity after upkeep venture", state.Val)
	}
}

func TestInitiativePassesOnEliminationToActivePlayer(t *testing.T) {
	g := mainPhaseGame(game.Player2)
	engine := NewEngine(nil)
	g.Players[game.Player1].HasInitiative = true

	engine.eliminatePlayer(g, game.Player1)

	if g.Players[game.Player1].HasInitiative {
		t.Fatal("eliminated player kept the initiative")
	}
	if !g.Players[game.Player2].HasInitiative {
		t.Fatal("initiative did not pass to the active player")
	}
}

func TestInitiativePassesOnEliminationSkippingLeaver(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	g.Players[game.Player1].HasInitiative = true

	// The active player (Player1) holds the initiative and leaves; it passes to
	// the next player in turn order.
	engine.eliminatePlayer(g, game.Player1)

	if g.Players[game.Player1].HasInitiative {
		t.Fatal("eliminated player kept the initiative")
	}
	if !g.Players[game.Player2].HasInitiative {
		t.Fatal("initiative did not pass to the next player in turn order")
	}
}

func TestLivingInitiativeIgnoresEliminatedHolder(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	g.Players[game.Player2].HasInitiative = true
	g.Players[game.Player2].Eliminated = true

	if livingInitiative(g).Exists {
		t.Fatal("livingInitiative returned an eliminated holder")
	}
	if !currentInitiative(g).Exists {
		t.Fatal("currentInitiative should still report the designation flag")
	}
}
