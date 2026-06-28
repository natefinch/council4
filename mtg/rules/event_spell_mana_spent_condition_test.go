package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/opt"
)

// TestEventSpellManaSpentToCastCondition proves the AggregateEventSpellManaSpentToCast
// comparison reads Event.ManaSpentToCast for the triggering cast and fails closed
// when the event is absent or carries no recorded spend, matching the cast-trigger
// intervening condition "if at least N mana was spent to cast it".
func TestEventSpellManaSpentToCastCondition(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	atLeastFour := opt.Val(game.Condition{
		Aggregates: []game.AggregateComparison{{
			Aggregate: game.AggregateEventSpellManaSpentToCast,
			Op:        compare.GreaterOrEqual,
			Value:     4,
		}},
	})
	noMana := opt.Val(game.Condition{
		Aggregates: []game.AggregateComparison{{
			Aggregate: game.AggregateEventSpellManaSpentToCast,
			Op:        compare.LessOrEqual,
			Value:     0,
		}},
	})

	// Fails closed when no triggering event is present.
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, atLeastFour) {
		t.Fatal("at-least-four satisfied with no event")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, noMana) {
		t.Fatal("no-mana satisfied with no event (must fail closed)")
	}

	// Fails closed when the event records no spend amount.
	missing := &game.Event{Kind: game.EventSpellCast, Controller: game.Player1}
	if conditionSatisfied(g, conditionContext{controller: game.Player1, event: missing}, atLeastFour) {
		t.Fatal("at-least-four satisfied when ManaSpentToCast absent")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player1, event: missing}, noMana) {
		t.Fatal("no-mana satisfied when ManaSpentToCast absent (must fail closed)")
	}

	cases := []struct {
		spent          int
		wantAtLeast    bool
		wantNoMana     bool
		describeAmount string
	}{
		{0, false, true, "zero"},
		{3, false, false, "three"},
		{4, true, false, "four"},
		{7, true, false, "seven"},
	}
	for _, tc := range cases {
		event := &game.Event{
			Kind:            game.EventSpellCast,
			Controller:      game.Player1,
			ManaSpentToCast: opt.Val(tc.spent),
		}
		ctx := conditionContext{controller: game.Player1, event: event}
		if got := conditionSatisfied(g, ctx, atLeastFour); got != tc.wantAtLeast {
			t.Errorf("at-least-four with %s mana = %v, want %v", tc.describeAmount, got, tc.wantAtLeast)
		}
		if got := conditionSatisfied(g, ctx, noMana); got != tc.wantNoMana {
			t.Errorf("no-mana with %s mana = %v, want %v", tc.describeAmount, got, tc.wantNoMana)
		}
	}
}
