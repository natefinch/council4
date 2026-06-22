package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WoodlandWanderer is the card definition for Woodland Wanderer.
//
// Type: Creature — Elemental
// Cost: {3}{G}
//
// Oracle text:
//
//	Vigilance, trample
//	Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
var WoodlandWanderer = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Woodland Wanderer",
		ManaCost: opt.Val(cost.Mana{
			cost.O(3),
			cost.G,
		}),
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
			game.TrampleStaticBody,
		},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersWithCountersReplacement("Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
				Kind:       game.DynamicAmountColorsOfManaSpentToCast,
				Multiplier: 1,
			})}),
		},
		OracleText: `
			Vigilance, trample
			Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
		`,
	},
}
