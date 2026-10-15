package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AdaptiveShimmerer is the card definition for Adaptive Shimmerer.
//
// Type: Creature — Insect
// Cost: {5}
//
// Oracle text:
//
//	Flash
//	This creature enters with three +1/+1 counters on it.
var AdaptiveShimmerer = newAdaptiveShimmerer

func newAdaptiveShimmerer() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Adaptive Shimmerer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with three +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
			},
			OracleText: `
			Flash
			This creature enters with three +1/+1 counters on it.
		`,
		},
	}
}
