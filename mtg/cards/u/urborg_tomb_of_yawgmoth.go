package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// UrborgTombOfYawgmoth is the card definition for Urborg, Tomb of Yawgmoth.
//
// Type: Legendary Land
//
// Oracle text:
//
//	Each land is a Swamp in addition to its other land types.
var UrborgTombOfYawgmoth = newUrborgTombOfYawgmoth

func newUrborgTombOfYawgmoth() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Urborg, Tomb of Yawgmoth",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each land is a Swamp in addition to its other land types.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}}),
							Layer:       game.LayerType,
							AddSubtypes: []types.Sub{types.Swamp},
						},
					},
				},
			},
			OracleText: `
			Each land is a Swamp in addition to its other land types.
		`,
		},
	}
}
