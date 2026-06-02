package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FuzzTargetEnumerationIsStable verifies that targetChoicesForSpecs returns
// identical results across multiple calls for the same board state.
// The seed corpus exercises 0–4 creatures across opposing players so that
// the combination-expansion code is exercised without being slow.
func FuzzTargetEnumerationIsStable(f *testing.F) {
	f.Add(0, 1) // 0 opponent creatures, 1 own creature
	f.Add(1, 0)
	f.Add(2, 1)
	f.Add(3, 2)
	f.Add(4, 4)

	f.Fuzz(func(t *testing.T, opponentCreatures, ownCreatures int) {
		// Clamp to a reasonable range to keep the test fast.
		if opponentCreatures < 0 {
			opponentCreatures = 0
		}
		if opponentCreatures > 6 {
			opponentCreatures = 6
		}
		if ownCreatures < 0 {
			ownCreatures = 0
		}
		if ownCreatures > 6 {
			ownCreatures = 6
		}

		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		for range opponentCreatures {
			addCombatPermanent(g, game.Player2, &game.CardDef{
				Name:      "Opponent Creature",
				Types:     []types.Card{types.Creature},
				Colors:    []color.Color{color.White},
				Power:     opt.Val(game.PT{Value: 2}),
				Toughness: opt.Val(game.PT{Value: 2}),
			})
		}
		for range ownCreatures {
			addCombatPermanent(g, game.Player1, &game.CardDef{
				Name:      "Own Creature",
				Types:     []types.Card{types.Creature},
				Colors:    []color.Color{color.Green},
				Power:     opt.Val(game.PT{Value: 1}),
				Toughness: opt.Val(game.PT{Value: 1}),
			})
		}

		specs := []game.TargetSpec{{MinTargets: 0, MaxTargets: 2, Constraint: "creature"}}

		first := targetChoicesForSpecs(g, game.Player1, nil, 0, specs)
		for i := range 5 {
			again := targetChoicesForSpecs(g, game.Player1, nil, 0, specs)
			if !reflect.DeepEqual(first.choices, again.choices) {
				t.Fatalf("iteration %d: choices changed - not deterministic", i+1)
			}
			if first.kind != again.kind {
				t.Fatalf("iteration %d: kind changed - not deterministic", i+1)
			}
		}
	})
}
