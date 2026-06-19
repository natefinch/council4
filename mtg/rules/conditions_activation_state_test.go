package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestConditionSourceCombatState verifies a source-bound "this creature is
// attacking" activation restriction is satisfied only while the source is
// declared as an attacker.
func TestConditionSourceCombatState(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Aggressive Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	condition := opt.Val(game.Condition{
		Object: opt.Val(game.SourcePermanentReference()),
		ObjectMatches: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			CombatState:   game.CombatStateAttacking,
		}),
	})
	ctx := conditionContext{controller: game.Player1, source: source}
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("attacking condition satisfied before any combat declaration")
	}
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("attacking condition not satisfied while source attacks")
	}
}

// TestConditionSourcePowerThreshold verifies a source-bound "this creature's
// power is N or greater" restriction tracks the source's power.
func TestConditionSourcePowerThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	weak := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Small Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	strong := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	}})
	condition := opt.Val(game.Condition{
		Object: opt.Val(game.SourcePermanentReference()),
		ObjectMatches: opt.Val(game.Selection{
			Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
		}),
	})
	if conditionSatisfied(g, conditionContext{controller: game.Player1, source: weak}, condition) {
		t.Fatal("power condition satisfied for a power-2 source")
	}
	if !conditionSatisfied(g, conditionContext{controller: game.Player1, source: strong}, condition) {
		t.Fatal("power condition not satisfied for a power-4 source")
	}
}

// TestConditionAnyOpponentPoisonAtLeast verifies an "an opponent has N or more
// poison counters" restriction reads opponents' poison totals.
func TestConditionAnyOpponentPoisonAtLeast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	condition := opt.Val(game.Condition{AnyOpponentPoisonAtLeast: 3})
	ctx := conditionContext{controller: game.Player1}
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("poison condition satisfied with no poison counters")
	}
	g.Players[game.Player2].PoisonCounters = 2
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("poison condition satisfied below threshold")
	}
	g.Players[game.Player2].PoisonCounters = 3
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("poison condition not satisfied at threshold")
	}
	// The controller's own poison must not satisfy an opponent-scoped condition.
	g.Players[game.Player2].PoisonCounters = 0
	g.Players[game.Player1].PoisonCounters = 5
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("poison condition satisfied by controller's own poison")
	}
}

// TestConditionControllerHandSizeExactly verifies an "exactly N cards in hand"
// restriction holds only at the precise hand size.
func TestConditionControllerHandSizeExactly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	condition := opt.Val(game.Condition{ControllerHandSizeExactly: opt.Val(7)})
	ctx := conditionContext{controller: game.Player1}
	for i := range 6 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  fmt.Sprintf("Hand Card %d", i),
			Types: []types.Card{types.Instant},
		}})
	}
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("exact-hand condition satisfied at six cards")
	}
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Hand Card 6",
		Types: []types.Card{types.Instant},
	}})
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("exact-hand condition not satisfied at seven cards")
	}
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Hand Card 7",
		Types: []types.Card{types.Instant},
	}})
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("exact-hand condition satisfied at eight cards")
	}
}

// TestConditionControlsCreatureWithKeyword verifies a "you control a creature
// with <keyword>" restriction matches a controlled creature carrying the
// keyword.
func TestConditionControlsCreatureWithKeyword(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	condition := opt.Val(game.Condition{
		ControlsMatching: opt.Val(game.SelectionCount{
			Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying},
		}),
	})
	ctx := conditionContext{controller: game.Player1}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grounded Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("keyword condition satisfied without a flying creature")
	}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Flier",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 1}),
		Toughness:       opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}})
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("keyword condition not satisfied with a flying creature")
	}
}
