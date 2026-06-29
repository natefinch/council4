package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestConditionSourceHasNoCountersOfKind verifies that the negated
// ObjectMatches + RequiredCounterCount condition correctly models "there are no
// depletion counters on this land". The condition is satisfied only when the
// source has zero counters of the named kind.
func TestConditionSourceHasNoCountersOfKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Depletion Land",
		Types: []types.Card{types.Land},
	}})
	source.Counters.Add(counter.Depletion, 2)

	noCounters := opt.Val(game.Condition{
		Object: opt.Val(game.SourcePermanentReference()),
		ObjectMatches: opt.Val(game.Selection{
			RequiredCounter:      counter.Depletion,
			RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 1}),
		}),
		Negate: true,
	})

	ctx := conditionContext{controller: game.Player1, source: source}

	if conditionSatisfied(g, ctx, noCounters) {
		t.Fatal("no-counters condition satisfied with 2 depletion counters on source")
	}

	source.Counters.Remove(counter.Depletion, 1)
	if conditionSatisfied(g, ctx, noCounters) {
		t.Fatal("no-counters condition satisfied with 1 depletion counter on source")
	}

	source.Counters.Remove(counter.Depletion, 1)
	if !conditionSatisfied(g, ctx, noCounters) {
		t.Fatal("no-counters condition not satisfied with 0 depletion counters on source")
	}
}
