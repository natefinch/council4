package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestGroupCantBlockThisTurnAllCreatures models the all-creatures form
// "Creatures can't block this turn." (Order // Chaos; Aragorn, King of Gondor's
// monarch clause): an object-less RuleEffectCantBlock with no controller scope
// and no selection filter stops every creature, of either player, from blocking
// for the turn, then expires at cleanup.
func TestGroupCantBlockThisTurnAllCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	theirs := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{Controller: game.Player1}

	resolveInstruction(engine, g, obj, game.ApplyRule{
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCantBlock,
			PermanentTypes: []types.Card{types.Creature},
		}},
		Duration: game.DurationThisTurn,
	}, nil)

	if canBlockWith(g, mine, game.Player1) {
		t.Fatal("all-creatures can't-block let the controller's creature block")
	}
	if canBlockWith(g, theirs, game.Player2) {
		t.Fatal("all-creatures can't-block let an opponent's creature block")
	}

	expireRuleEffects(g)

	if !canBlockWith(g, theirs, game.Player2) {
		t.Fatal("all-creatures can't-block still applied after cleanup expiry")
	}
}
