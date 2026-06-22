package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TajuruStalwart is the card definition for Tajuru Stalwart.
//
// Type: Creature — Elf Scout Ally
// Cost: {2}{G}
//
// Oracle text:
//
//	Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
var TajuruStalwart = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Tajuru Stalwart",
		ManaCost: opt.Val(cost.Mana{
			cost.O(2),
			cost.G,
		}),
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf, types.Scout, types.Ally},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersWithCountersReplacement("Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
				Kind:       game.DynamicAmountColorsOfManaSpentToCast,
				Multiplier: 1,
			})}),
		},
		OracleText: `
			Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
		`,
	},
}
