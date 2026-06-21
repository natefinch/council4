package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestContinuousEffectCounterFilteredGroupGrantsKeywordOnlyToCounterHolders
// proves the counter-matters anthem static ("Each creature you control with a
// +1/+1 counter on it has flying") grants the keyword only to creatures that
// currently bear a +1/+1 counter.
func TestContinuousEffectCounterFilteredGroupGrantsKeywordOnlyToCounterHolders(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	withCounter := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	withCounter.Counters.Add(counter.PlusOnePlusOne, 1)
	withoutCounter := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: withCounter.ObjectID,
		Layer:          game.LayerAbility,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes:   []types.Card{types.Creature},
			MatchCounter:    true,
			RequiredCounter: counter.PlusOnePlusOne,
		}),
		AddKeywords: []game.Keyword{game.Flying},
	})

	if !hasKeyword(g, withCounter, game.Flying) {
		t.Fatal("creature with a +1/+1 counter should have flying")
	}
	if hasKeyword(g, withoutCounter, game.Flying) {
		t.Fatal("creature without a +1/+1 counter should not have flying")
	}
}

// TestContinuousEffectAnyCounterFilteredGroupGrantsAbilityOnlyToCounterHolders
// proves the kind-agnostic counter-matters static ("Each creature you control
// with a counter on it has ...", Rishkar) grants only to your creatures that
// currently bear a counter of any kind: a creature with a non-+1/+1 counter
// qualifies, one without a counter does not, and an opponent's counter-bearing
// creature is excluded by the "you control" scope.
func TestContinuousEffectAnyCounterFilteredGroupGrantsAbilityOnlyToCounterHolders(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	withChargeCounter := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	withChargeCounter.Counters.Add(counter.Charge, 1)
	withoutCounter := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opponentWithCounter := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	opponentWithCounter.Counters.Add(counter.PlusOnePlusOne, 1)

	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: withChargeCounter.ObjectID,
		Layer:          game.LayerAbility,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes:   []types.Card{types.Creature},
			MatchAnyCounter: true,
		}),
		AddKeywords: []game.Keyword{game.Flying},
	})

	if !hasKeyword(g, withChargeCounter, game.Flying) {
		t.Fatal("your creature with any counter should have flying")
	}
	if hasKeyword(g, withoutCounter, game.Flying) {
		t.Fatal("your creature without a counter should not have flying")
	}
	if hasKeyword(g, opponentWithCounter, game.Flying) {
		t.Fatal("opponent's counter-bearing creature should not have flying")
	}
}
