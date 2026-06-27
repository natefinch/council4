package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestSelectionDealtDamageThisTurnFilter verifies that a Selection carrying the
// DealtDamageThisTurn target filter matches only permanents that received damage
// this turn. A creature that was dealt damage matches; one that was not is
// excluded, even though both are creatures controlled by the same player.
func TestSelectionDealtDamageThisTurnFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	damaged := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Wounded Bear",
		Types: []types.Card{types.Creature},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Unharmed Bear",
		Types: []types.Card{types.Creature},
	}})

	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		DamageRecipient: game.DamageRecipientPermanent,
		PermanentID:     damaged.ObjectID,
	})

	ctx := conditionContext{controller: game.Player1}

	atLeastOne := opt.Val(game.Condition{ControlsMatching: opt.Val(game.SelectionCount{
		MinCount:  1,
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, DealtDamageThisTurn: true},
	})})
	if !conditionSatisfied(g, ctx, atLeastOne) {
		t.Fatal("dealt-damage-this-turn filter did not match the damaged creature")
	}

	atLeastTwo := opt.Val(game.Condition{ControlsMatching: opt.Val(game.SelectionCount{
		MinCount:  2,
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, DealtDamageThisTurn: true},
	})})
	if conditionSatisfied(g, ctx, atLeastTwo) {
		t.Fatal("dealt-damage-this-turn filter matched the undamaged creature")
	}
}
