package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RisingOfTheDay is the card definition for Rising of the Day.
//
// Type: Enchantment
// Cost: {2}{R}
//
// Oracle text:
//
//	Creatures you control have haste.
//	Legendary creatures you control get +1/+0.
var RisingOfTheDay = newRisingOfTheDay

func newRisingOfTheDay() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Rising of the Day",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control have haste.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Haste},
						},
					},
				},
				game.StaticAbility{
					Text: "Legendary creatures you control get +1/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Creatures you control have haste.
			Legendary creatures you control get +1/+0.
		`,
		},
	}
}
