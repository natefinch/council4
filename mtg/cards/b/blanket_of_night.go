package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlanketOfNight is the card definition for Blanket of Night.
//
// Type: Enchantment
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Each land is a Swamp in addition to its other land types.
var BlanketOfNight = newBlanketOfNight

func newBlanketOfNight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Blanket of Night",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
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
