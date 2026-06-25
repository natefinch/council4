package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func greatestToughnessCreature(g *game.Game, controller game.PlayerID, name string, toughness int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}})
}

// TestConditionControllerControlsGreatestToughnessCreature exercises the Abzan
// Beastmaster predicate "you control the creature with the greatest toughness or
// tied for the greatest toughness." It holds when the controller owns the sole
// highest-toughness creature or is tied for highest, and fails when an
// opponent's creature is strictly tougher or when no creatures exist.
func TestConditionControllerControlsGreatestToughnessCreature(t *testing.T) {
	condition := opt.Val(game.Condition{ControllerControlsGreatestToughnessCreature: true})

	t.Run("no creatures", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		if conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
			t.Fatal("condition held with no creatures on the battlefield")
		}
	})

	t.Run("sole greatest", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		greatestToughnessCreature(g, game.Player1, "Mine", 5)
		greatestToughnessCreature(g, game.Player2, "Theirs", 3)
		if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
			t.Fatal("condition did not hold for the sole greatest-toughness creature")
		}
		if conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
			t.Fatal("condition held for the less-tough controller")
		}
	})

	t.Run("tied for greatest", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		greatestToughnessCreature(g, game.Player1, "Mine", 4)
		greatestToughnessCreature(g, game.Player2, "Theirs", 4)
		if !conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
			t.Fatal("condition did not hold for a tie for greatest toughness")
		}
		if !conditionSatisfied(g, conditionContext{controller: game.Player2}, condition) {
			t.Fatal("condition did not hold for the other tied controller")
		}
	})

	t.Run("opponent strictly greater", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		greatestToughnessCreature(g, game.Player1, "Mine", 2)
		greatestToughnessCreature(g, game.Player1, "MineAlso", 3)
		greatestToughnessCreature(g, game.Player2, "Theirs", 6)
		if conditionSatisfied(g, conditionContext{controller: game.Player1}, condition) {
			t.Fatal("condition held while an opponent's creature was strictly tougher")
		}
	})
}
