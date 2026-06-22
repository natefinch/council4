package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkyriderElf is the card definition for Skyrider Elf.
//
// Type: Creature — Elf Warrior Ally
// Cost: {X}{G}{U}
//
// Oracle text:
//
//	Flying
//	Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
var SkyriderElf = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Green),
	CardFace: game.CardFace{
		Name: "Skyrider Elf",
		ManaCost: opt.Val(cost.Mana{
			cost.X,
			cost.G,
			cost.U,
		}),
		Colors:    []color.Color{color.Green, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf, types.Warrior, types.Ally},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersWithCountersReplacement("Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
				Kind:       game.DynamicAmountColorsOfManaSpentToCast,
				Multiplier: 1,
			})}),
		},
		OracleText: `
			Flying
			Converge — This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
		`,
	},
}
