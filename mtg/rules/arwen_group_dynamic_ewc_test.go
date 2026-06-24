package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestGroupDynamicEntersWithCountersResolvesAgainstSource verifies that Arwen,
// Weaver of Hope's group enters-with-counters replacement ("Each other creature
// you control enters with a number of additional +1/+1 counters on it equal to
// Arwen's toughness.") evaluates its dynamic amount against the SOURCE (Arwen),
// not the entering permanent. Other controlled creatures must enter with
// Arwen's-toughness-many +1/+1 counters regardless of their own toughness.
func TestGroupDynamicEntersWithCountersResolvesAgainstSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	arwen := &game.CardDef{CardFace: game.CardFace{
		Name:      "Arwen, Weaver of Hope",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 3}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersWithCountersGroupReplacement(
				"Each other creature you control enters with a number of additional +1/+1 counters on it equal to Arwen's toughness.",
				&game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerYou,
					ExcludeSource: true,
				},
				game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountObjectToughness,
					Multiplier: 1,
					Object:     game.SourcePermanentReference(),
				})},
			),
		},
	}}
	sourcePerm := addReplacementPermanent(t, g, game.Player1, arwen)
	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("registered replacement effects = %d, want 1", len(g.ReplacementEffects))
	}
	if got := sourcePerm.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("source +1/+1 counters = %d, want 0 (ExcludeSource)", got)
	}

	// A small creature whose own toughness is 1 must still enter with 3 counters
	// (Arwen's toughness), proving the dynamic amount resolves against the source.
	bear := &game.CardDef{CardFace: game.CardFace{
		Name:      "Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
	ownCreature := addReplacementPermanent(t, g, game.Player1, bear)
	if got := ownCreature.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("controlled creature +1/+1 counters = %d, want 3 (Arwen's toughness)", got)
	}

	opponentCreature := addReplacementPermanent(t, g, game.Player2, bear)
	if got := opponentCreature.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("opponent creature +1/+1 counters = %d, want 0", got)
	}
}
