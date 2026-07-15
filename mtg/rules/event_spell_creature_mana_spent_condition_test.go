package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/opt"
)

// TestEventSpellCreatureManaSpentToCastCondition proves the
// AggregateEventSpellManaFromCreaturesSpentToCast comparison reads
// Event.ManaFromCreaturesSpentToCast for the triggering cast and fails closed
// when the event is absent or carries no recorded creature-mana spend. This is
// Inga and Esika's intervening condition "if three or more mana from creatures
// was spent to cast it, draw a card": the 2-vs-3 threshold is the whole point.
func TestEventSpellCreatureManaSpentToCastCondition(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	atLeastThree := opt.Val(game.Condition{
		Aggregates: []game.AggregateComparison{{
			Aggregate: game.AggregateEventSpellManaFromCreaturesSpentToCast,
			Op:        compare.GreaterOrEqual,
			Value:     3,
		}},
	})

	// Fails closed when no triggering event is present.
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, atLeastThree) {
		t.Fatal("at-least-three satisfied with no event")
	}

	// Fails closed when the event records no creature-mana spend amount.
	missing := &game.Event{Kind: game.EventSpellCast, Controller: game.Player1}
	if conditionSatisfied(g, conditionContext{controller: game.Player1, event: missing}, atLeastThree) {
		t.Fatal("at-least-three satisfied when ManaFromCreaturesSpentToCast absent")
	}

	cases := []struct {
		creatureMana int
		want         bool
		describe     string
	}{
		{0, false, "zero"},
		{2, false, "exactly two"},
		{3, true, "exactly three"},
		{5, true, "five"},
	}
	for _, tc := range cases {
		event := &game.Event{
			Kind:                         game.EventSpellCast,
			Controller:                   game.Player1,
			ManaFromCreaturesSpentToCast: opt.Val(tc.creatureMana),
		}
		ctx := conditionContext{controller: game.Player1, event: event}
		if got := conditionSatisfied(g, ctx, atLeastThree); got != tc.want {
			t.Errorf("at-least-three with %s creature mana = %v, want %v", tc.describe, got, tc.want)
		}
	}
}
