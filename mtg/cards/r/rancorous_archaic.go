package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RancorousArchaic is the card definition for Rancorous Archaic.
//
// Type: Creature — Avatar
// Cost: {5}
//
// Oracle text:
//
//	Trample, reach
//	Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
var RancorousArchaic = func() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Rancorous Archaic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Avatar},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.ReachStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountColorsOfManaSpentToCast,
					Multiplier: 1,
				})}),
			},
			OracleText: `
			Trample, reach
			Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
		`,
		},
	}
}
