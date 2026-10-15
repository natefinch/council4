package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AbsoluteGrace is the card definition for Absolute Grace.
//
// Type: Enchantment
// Cost: {1}{W}
//
// Oracle text:
//
//	All creatures have protection from black.
var AbsoluteGrace = newAbsoluteGrace

func newAbsoluteGrace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Absolute Grace",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All creatures have protection from black.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.ProtectionFromColorsStaticAbility(color.Black)),
							},
						},
					},
				},
			},
			OracleText: `
			All creatures have protection from black.
		`,
		},
	}
}
