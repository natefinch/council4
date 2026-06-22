package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestConditionSourceSaddled verifies the "this creature is saddled" /
// "isn't saddled" source-state predicate tracks the source Mount's saddled
// flag, including the negated form used by the "Otherwise" branch of the
// Caustic Bronco saddled conditional.
func TestConditionSourceSaddled(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Test Mount",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	ctx := conditionContext{controller: game.Player1, source: source}

	saddled := opt.Val(game.Condition{SourceSaddled: true})
	notSaddled := opt.Val(game.Condition{SourceSaddled: true, Negate: true})

	if conditionSatisfied(g, ctx, saddled) {
		t.Fatal("saddled condition satisfied before the Mount is saddled")
	}
	if !conditionSatisfied(g, ctx, notSaddled) {
		t.Fatal("not-saddled condition unsatisfied before the Mount is saddled")
	}

	source.Saddled = true

	if !conditionSatisfied(g, ctx, saddled) {
		t.Fatal("saddled condition unsatisfied after the Mount is saddled")
	}
	if conditionSatisfied(g, ctx, notSaddled) {
		t.Fatal("not-saddled condition satisfied after the Mount is saddled")
	}
}
