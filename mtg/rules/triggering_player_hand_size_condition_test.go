package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestTriggeringPlayerHandSizeCondition proves the AggregateEventPlayerHandSize
// comparison reads the triggering step event's player, not the ability
// controller, and fails closed when no event is present, backing the phase
// trigger "if that player has N or fewer/more cards in hand".
func TestTriggeringPlayerHandSizeCondition(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	atMostTwo := opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{
		Aggregate: game.AggregateEventPlayerHandSize, Op: compare.LessOrEqual, Value: 2,
	}}})
	atLeastFive := opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{
		Aggregate: game.AggregateEventPlayerHandSize, Op: compare.GreaterOrEqual, Value: 5,
	}}})

	if conditionSatisfied(g, conditionContext{controller: game.Player1}, atMostTwo) {
		t.Fatal("at-most-two satisfied with no event (must fail closed)")
	}

	event := &game.Event{Kind: game.EventBeginningOfStep, Controller: game.Player2, Player: game.Player2, Step: game.StepUpkeep}
	if !conditionSatisfied(g, conditionContext{controller: game.Player1, event: event}, atMostTwo) {
		t.Fatal("at-most-two not satisfied for empty-handed triggering player")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player1, event: event}, atLeastFive) {
		t.Fatal("at-least-five satisfied for empty-handed triggering player")
	}

	for range 5 {
		addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Card", Types: []types.Card{types.Instant}}})
	}
	if !conditionSatisfied(g, conditionContext{controller: game.Player1, event: event}, atLeastFive) {
		t.Fatal("at-least-five not satisfied for five-card triggering player")
	}
	if conditionSatisfied(g, conditionContext{controller: game.Player1, event: event}, atMostTwo) {
		t.Fatal("at-most-two satisfied for five-card triggering player")
	}
}
