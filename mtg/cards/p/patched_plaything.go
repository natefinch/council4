package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PatchedPlaything is the card definition for Patched Plaything.
//
// Type: Artifact Creature — Toy
// Cost: {2}{W}
//
// Oracle text:
//
//	Double strike
//	This creature enters with two -1/-1 counters on it if you cast it from your hand.
var PatchedPlaything = newPatchedPlaything

func newPatchedPlaything() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Patched Plaything",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Toy},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("This creature enters with two -1/-1 counters on it if you cast it from your hand.", &game.Condition{
					EventPermanentWasCastFromControllerHand: true,
				}, game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 2}),
			},
			OracleText: `
			Double strike
			This creature enters with two -1/-1 counters on it if you cast it from your hand.
		`,
		},
	}
}
