package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrideOfThePerfect is the card definition for Pride of the Perfect.
//
// Type: Enchantment
// Cost: {3}{B}
//
// Oracle text:
//
//	Elves you control get +2/+0.
var PrideOfThePerfect = newPrideOfThePerfect

func newPrideOfThePerfect() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Pride of the Perfect",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Elves you control get +2/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 2,
						},
					},
				},
			},
			OracleText: `
			Elves you control get +2/+0.
		`,
		},
	}
}
