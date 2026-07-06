package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestConditionControllerDesignations exercises the controller-designation
// intervening-if predicates that read the ability controller's live monarch,
// initiative, and city's-blessing flags ("if you're the monarch", "if you have
// the initiative", "if you have the city's blessing"). Each holds only when the
// context controller holds the matching designation.
func TestConditionControllerDesignations(t *testing.T) {
	tests := []struct {
		name      string
		condition game.Condition
		grant     func(*game.Player)
	}{
		{
			name:      "monarch",
			condition: game.Condition{ControllerIsMonarch: true},
			grant:     func(p *game.Player) { p.IsMonarch = true },
		},
		{
			name:      "initiative",
			condition: game.Condition{ControllerHasInitiative: true},
			grant:     func(p *game.Player) { p.HasInitiative = true },
		},
		{
			name:      "city's blessing",
			condition: game.Condition{ControllerHasCityBlessing: true},
			grant:     func(p *game.Player) { p.HasCityBlessing = true },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			condition := opt.Val(tc.condition)

			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
				t.Fatal("condition held before the designation was granted")
			}

			g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
			tc.grant(g.Players[0])
			if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
				t.Fatal("condition did not hold for the designated controller")
			}
			if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
				t.Fatal("condition held for a controller without the designation")
			}
		})
	}
}

// TestConditionNoMonarch exercises the "if there is no monarch" intervening-if
// predicate (Crown of Gondor, Archivist of Gondor). It holds only while no player
// currently holds the monarch designation, independent of the context controller.
func TestConditionNoMonarch(t *testing.T) {
	condition := opt.Val(game.Condition{NoMonarch: true})

	// A fresh game has no monarch, so the condition holds for any controller.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition did not hold when there was no monarch")
	}
	if !conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("condition did not hold for another controller when there was no monarch")
	}

	// Once any player is the monarch, the condition holds for nobody.
	g.Players[game.Player2].IsMonarch = true
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition held while a player was the monarch")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("condition held for the monarch themselves")
	}

	// A monarch who has left the game no longer counts (their IsMonarch flag is
	// not cleared on elimination), so the crown is vacant again.
	g.Players[game.Player2].Eliminated = true
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition did not hold after the monarch left the game")
	}
}

// TestConditionAnOpponentIsMonarch exercises the opponent-designation
// intervening-if predicate "if an opponent is the monarch" (Queen Marchesa). It
// holds only when a player other than the context controller currently holds the
// monarch designation.
func TestConditionAnOpponentIsMonarch(t *testing.T) {
	condition := opt.Val(game.Condition{AnOpponentIsMonarch: true})

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition held before any opponent became the monarch")
	}

	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player2].IsMonarch = true
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition did not hold when the controller's opponent was the monarch")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("condition held when the controller themselves was the monarch")
	}
}

// TestConditionControllerWasMonarchAtTurnStart exercises the turn-start monarch
// snapshot predicate "if you were the monarch as the turn began" (Knights of the
// Black Rose). It reads Turn.MonarchAtTurnStart, not the live designation, so it
// holds for the player recorded as the monarch when the turn advanced even after
// the crown changes hands mid-turn.
func TestConditionControllerWasMonarchAtTurnStart(t *testing.T) {
	condition := opt.Val(game.Condition{ControllerWasMonarchAtTurnStart: true})

	// No snapshot: the condition holds for nobody.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition held with no turn-start monarch snapshot")
	}

	// Player1 was the monarch as the turn began; the crown then passed to Player2.
	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.MonarchAtTurnStart = opt.Val(game.Player1)
	g.Players[game.Player2].IsMonarch = true
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
		t.Fatal("condition did not hold for the player who was the monarch at turn start")
	}
	// The current (mid-turn) monarch was not the monarch at turn start.
	if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
		t.Fatal("condition held for the current monarch who was not the turn-start monarch")
	}
}
