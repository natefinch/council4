package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarkestHour is the card definition for Darkest Hour.
//
// Type: Enchantment
// Cost: {B}
//
// Oracle text:
//
//	All creatures are black.
var DarkestHour = newDarkestHour

func newDarkestHour() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Darkest Hour",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All creatures are black.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:  game.DurationPermanent,
							Group:     game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:     game.LayerColor,
							SetColors: []color.Color{color.Black},
						},
					},
				},
			},
			OracleText: `
			All creatures are black.
		`,
		},
	}
}
