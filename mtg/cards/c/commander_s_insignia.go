package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CommanderSInsignia is the card definition for Commander's Insignia.
//
// Type: Enchantment
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Creatures you control get +1/+1 for each time you've cast your commander from the command zone this game.
var CommanderSInsignia = newCommanderSInsignia()

func newCommanderSInsignia() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Commander's Insignia",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerPowerToughnessModify,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCommanderCastCount,
								Multiplier: 1,
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCommanderCastCount,
								Multiplier: 1,
							}),
						},
					},
				},
			},
			OracleText: `
			Creatures you control get +1/+1 for each time you've cast your commander from the command zone this game.
		`,
		},
	}
}
