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
