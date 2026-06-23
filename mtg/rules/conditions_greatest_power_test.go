package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func greatestPowerCreature(g *game.Game, controller game.PlayerID, name string, power int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: power}),
	}})
}

// TestConditionControllerControlsGreatestPowerCreature exercises the Summon:
// Fenrir chapter III predicate "you control the creature with the greatest power
// or tied for the greatest power." It holds when the controller owns the sole
// highest-power creature or is tied for highest, and fails when an opponent's
// creature is strictly more powerful or when no creatures exist.
func TestConditionControllerControlsGreatestPowerCreature(t *testing.T) {
	condition := opt.Val(game.Condition{ControllerControlsGreatestPowerCreature: true})

	t.Run("no creatures", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		if conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
			t.Fatal("condition held with no creatures on the battlefield")
		}
	})

	t.Run("sole greatest", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		greatestPowerCreature(g, game.Player1, "Mine", 5)
		greatestPowerCreature(g, game.Player2, "Theirs", 3)
		if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
			t.Fatal("condition did not hold for the sole greatest-power creature")
		}
		if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
			t.Fatal("condition held for the weaker controller")
		}
	})

	t.Run("tied for greatest", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		greatestPowerCreature(g, game.Player1, "Mine", 4)
		greatestPowerCreature(g, game.Player2, "Theirs", 4)
		if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
			t.Fatal("condition did not hold for a tie for greatest power")
		}
		if !conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
			t.Fatal("condition did not hold for the other tied controller")
		}
	})

	t.Run("opponent strictly greater", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		greatestPowerCreature(g, game.Player1, "Mine", 2)
		greatestPowerCreature(g, game.Player1, "MineAlso", 3)
		greatestPowerCreature(g, game.Player2, "Theirs", 6)
		if conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
			t.Fatal("condition held while an opponent's creature was strictly more powerful")
		}
	})
}
