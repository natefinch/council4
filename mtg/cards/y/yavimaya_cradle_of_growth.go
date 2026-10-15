package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// YavimayaCradleOfGrowth is the card definition for Yavimaya, Cradle of Growth.
//
// Type: Legendary Land
//
// Oracle text:
//
//	Each land is a Forest in addition to its other land types.
var YavimayaCradleOfGrowth = newYavimayaCradleOfGrowth

func newYavimayaCradleOfGrowth() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Yavimaya, Cradle of Growth",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each land is a Forest in addition to its other land types.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}}),
							Layer:       game.LayerType,
							AddSubtypes: []types.Sub{types.Forest},
						},
					},
				},
			},
			OracleText: `
			Each land is a Forest in addition to its other land types.
		`,
		},
	}
}
