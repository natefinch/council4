package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrismaticOmen is the card definition for Prismatic Omen.
//
// Type: Enchantment
// Cost: {1}{G}
//
// Oracle text:
//
//	Lands you control are every basic land type in addition to their other types.
var PrismaticOmen = newPrismaticOmen

func newPrismaticOmen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Prismatic Omen",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Lands you control are every basic land type in addition to their other types.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:              game.DurationPermanent,
							Group:                 game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Land}}),
							Layer:                 game.LayerType,
							AddEveryBasicLandType: true,
						},
					},
				},
			},
			OracleText: `
			Lands you control are every basic land type in addition to their other types.
		`,
		},
	}
}
