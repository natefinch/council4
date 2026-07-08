package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpringmantleCleric is the card definition for Springmantle Cleric.
//
// Type: Creature — Elf Cleric
// Cost: {4}{G}
//
// Oracle text:
//
//	This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
var SpringmantleCleric = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Springmantle Cleric",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountColorsOfManaSpentToCast,
					Multiplier: 1,
				})}),
			},
			OracleText: `
			This creature enters with a +1/+1 counter on it for each color of mana spent to cast it.
		`,
		},
	}
}
