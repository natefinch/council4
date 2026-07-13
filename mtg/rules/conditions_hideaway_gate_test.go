package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/opt"
)

// anyOpponentDamageCondition builds the Spinerock Knoll Hideaway gate: an
// opponent was dealt threshold or more damage this turn.
func anyOpponentDamageCondition(threshold int) opt.V[game.Condition] {
	return opt.Val(game.Condition{
		Aggregates: []game.AggregateComparison{{
			Aggregate: game.AggregateAnyOpponentDamageTakenThisTurn,
			Op:        compare.GreaterOrEqual,
			Value:     threshold,
		}},
	})
}

// TestConditionAnyOpponentDealtDamageThisTurn covers Spinerock Knoll's gate:
// only player-recipient damage dealt to an opponent this turn counts, and it is
// summed per opponent before the inclusive threshold is applied.
func TestConditionAnyOpponentDealtDamageThisTurn(t *testing.T) {
	g := buildTwoTurnEventLog(
		[]game.Event{
			// Prior-turn damage must not count toward this turn's total.
			{Kind: game.EventDamageDealt, DamageRecipient: game.DamageRecipientPlayer, Player: game.Player2, Amount: 9},
		},
		[]game.Event{
			{Kind: game.EventDamageDealt, DamageRecipient: game.DamageRecipientPlayer, Player: game.Player2, Amount: 4},
			{Kind: game.EventDamageDealt, DamageRecipient: game.DamageRecipientPlayer, Player: game.Player2, Amount: 3},
			// Damage to the controller and to a permanent is irrelevant.
			{Kind: game.EventDamageDealt, DamageRecipient: game.DamageRecipientPlayer, Player: game.Player1, Amount: 20},
			{Kind: game.EventDamageDealt, DamageRecipient: game.DamageRecipientPermanent, Player: game.Player2, Amount: 20},
		},
	)
	ctx := conditionContext{controller: game.Player1}

	if !conditionSatisfied(g, ctx, anyOpponentDamageCondition(7)) {
		t.Fatal("opponent dealt 7 this turn should satisfy the >= 7 gate")
	}
	if conditionSatisfied(g, ctx, anyOpponentDamageCondition(8)) {
		t.Fatal("opponent dealt only 7 this turn must not satisfy the >= 8 gate")
	}
}

// TestConditionAnyOpponentDamageIgnoresControllerDamage confirms the predicate
// is existential over opponents only: damage dealt solely to the controller
// never unlocks the gate.
func TestConditionAnyOpponentDamageIgnoresControllerDamage(t *testing.T) {
	g := buildTwoTurnEventLog(nil, []game.Event{
		{Kind: game.EventDamageDealt, DamageRecipient: game.DamageRecipientPlayer, Player: game.Player1, Amount: 15},
	})
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, anyOpponentDamageCondition(7)) {
		t.Fatal("damage dealt only to the controller must not satisfy an opponent-damage gate")
	}
}

// TestConditionAnyLibrarySizeAtMost covers Shelldock Isle's gate: "a library"
// means any player's library, so the smallest library size decides the gate.
func TestConditionAnyLibrarySizeAtMost(t *testing.T) {
	condition := opt.Val(game.Condition{
		Aggregates: []game.AggregateComparison{{
			Aggregate: game.AggregateMinPlayerLibrarySize,
			Op:        compare.LessOrEqual,
			Value:     20,
		}},
	})
	ctx := conditionContext{controller: game.Player1}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Give every player a large library so no library is at or below 20.
	for playerID := range game.PlayerID(game.NumPlayers) {
		for range 21 {
			addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "Filler"}})
		}
	}
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("all libraries above 20 must not satisfy the <= 20 gate")
	}

	// Draining one player's library to 20 satisfies the existential gate.
	only := g.Players[game.Player3].Library.All()[0]
	g.Players[game.Player3].Library.Remove(only)
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("a single library at 20 should satisfy the <= 20 gate")
	}
}

// TestConditionAllPlayersHandEmpty covers Howltooth Hollow's universal gate:
// every player must have an empty hand.
func TestConditionAllPlayersHandEmpty(t *testing.T) {
	condition := opt.Val(game.Condition{AllPlayersHandEmpty: true})
	ctx := conditionContext{controller: game.Player1}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("empty hands for every player should satisfy the gate")
	}

	addCardToHand(g, game.Player4, &game.CardDef{CardFace: game.CardFace{Name: "Held Card"}})
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("a single non-empty hand must fail the each-player-empty gate")
	}
}
