package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// manaValueArtifact adds an artifact the given player controls whose mana value
// is manaValue (a bare generic cost, or no cost at all for a zero mana value so
// it models an artifact token).
func manaValueArtifact(g *game.Game, controller game.PlayerID, name string, manaValue int) *game.Permanent {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Artifact},
	}}
	if manaValue > 0 {
		def.ManaCost = opt.Val(cost.Mana{cost.O(manaValue)})
	}
	return addCombatPermanent(g, controller, def)
}

// manaValueCreature adds a non-artifact creature the given player controls whose
// mana value is manaValue, used to confirm the artifact filter ignores it.
func manaValueCreature(g *game.Game, controller game.PlayerID, name string, manaValue int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.O(manaValue)}),
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
}

// TestConditionControlsGreatestManaValueInGroup exercises Padeem, Consul of
// Innovation's intervening-if "you control the artifact with the greatest mana
// value or tied for the greatest mana value." It holds when the controller owns
// the sole highest-mana-value artifact or is tied for highest, requires the
// controller to actually control a matching artifact, and reads the group filter
// so non-artifacts never count.
func TestConditionControlsGreatestManaValueInGroup(t *testing.T) {
	artifactCondition := opt.Val(game.Condition{
		ControlsGreatestManaValueInGroup: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
	})

	t.Run("no artifacts", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		if conditionSatisfied(g, conditionContext{controller: game.Player1}, artifactCondition) {
			t.Fatal("condition held with no artifacts on the battlefield")
		}
	})

	t.Run("sole greatest", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		manaValueArtifact(g, game.Player1, "Mine", 5)
		manaValueArtifact(g, game.Player2, "Theirs", 3)
		if !conditionSatisfied(g, conditionContext{controller: game.Player1}, artifactCondition) {
			t.Fatal("condition did not hold for the sole greatest-mana-value artifact")
		}
		if conditionSatisfied(g, conditionContext{controller: game.Player2}, artifactCondition) {
			t.Fatal("condition held for the lower-mana-value controller")
		}
	})

	t.Run("tied for greatest", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		manaValueArtifact(g, game.Player1, "Mine", 4)
		manaValueArtifact(g, game.Player2, "Theirs", 4)
		if !conditionSatisfied(g, conditionContext{controller: game.Player1}, artifactCondition) {
			t.Fatal("condition did not hold for a tie for greatest mana value")
		}
		if !conditionSatisfied(g, conditionContext{controller: game.Player2}, artifactCondition) {
			t.Fatal("condition did not hold for the other tied controller")
		}
	})

	t.Run("opponent strictly greater", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		manaValueArtifact(g, game.Player1, "Mine", 2)
		manaValueArtifact(g, game.Player1, "MineAlso", 3)
		manaValueArtifact(g, game.Player2, "Theirs", 6)
		if conditionSatisfied(g, conditionContext{controller: game.Player1}, artifactCondition) {
			t.Fatal("condition held while an opponent's artifact had strictly greater mana value")
		}
	})

	t.Run("controls no matching artifact", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		manaValueCreature(g, game.Player1, "MyCreature", 9)
		manaValueArtifact(g, game.Player2, "TheirArtifact", 1)
		if conditionSatisfied(g, conditionContext{controller: game.Player1}, artifactCondition) {
			t.Fatal("condition held while the controller controlled no matching artifact")
		}
		if !conditionSatisfied(g, conditionContext{controller: game.Player2}, artifactCondition) {
			t.Fatal("condition did not hold for the sole artifact controller")
		}
	})

	t.Run("zero mana value artifact tokens tie", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		manaValueArtifact(g, game.Player1, "MyToken", 0)
		manaValueArtifact(g, game.Player2, "TheirToken", 0)
		if !conditionSatisfied(g, conditionContext{controller: game.Player1}, artifactCondition) {
			t.Fatal("condition did not hold for a zero-mana-value tie")
		}
	})

	t.Run("greater non-artifact ignored", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		manaValueArtifact(g, game.Player1, "MyArtifact", 0)
		manaValueCreature(g, game.Player2, "TheirBigCreature", 9)
		if !conditionSatisfied(g, conditionContext{controller: game.Player1}, artifactCondition) {
			t.Fatal("condition did not hold: a non-artifact with greater mana value must not count")
		}
	})

	t.Run("phased-out artifact ignored", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		manaValueArtifact(g, game.Player1, "Mine", 2)
		phased := manaValueArtifact(g, game.Player2, "TheirPhased", 9)
		phased.PhasedOut = true
		if !conditionSatisfied(g, conditionContext{controller: game.Player1}, artifactCondition) {
			t.Fatal("condition did not hold while the greater opposing artifact was phased out")
		}
	})
}

// TestConditionControlsGreatestManaValueInGroupGeneric confirms the predicate
// generalizes past artifacts: with a creature filter it evaluates the greatest
// mana value among creatures, ignoring artifacts entirely.
func TestConditionControlsGreatestManaValueInGroupGeneric(t *testing.T) {
	creatureCondition := opt.Val(game.Condition{
		ControlsGreatestManaValueInGroup: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	})

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	manaValueCreature(g, game.Player1, "MyCreature", 5)
	manaValueArtifact(g, game.Player2, "TheirBigArtifact", 9)
	manaValueCreature(g, game.Player2, "TheirCreature", 3)
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, creatureCondition) {
		t.Fatal("creature-filtered condition did not hold for the greatest-mana-value creature controller")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player2}, creatureCondition) {
		t.Fatal("creature-filtered condition held for the lower-mana-value creature controller")
	}
}
