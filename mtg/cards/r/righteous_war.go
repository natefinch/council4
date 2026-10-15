package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RighteousWar is the card definition for Righteous War.
//
// Type: Enchantment
// Cost: {1}{W}{B}
//
// Oracle text:
//
//	White creatures you control have protection from black.
//	Black creatures you control have protection from white.
var RighteousWar = newRighteousWar

func newRighteousWar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Righteous War",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "White creatures you control have protection from black.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.White}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.ProtectionFromColorsStaticAbility(color.Black)),
							},
						},
					},
				},
				game.StaticAbility{
					Text: "Black creatures you control have protection from white.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Black}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.ProtectionFromColorsStaticAbility(color.White)),
							},
						},
					},
				},
			},
			OracleText: `
			White creatures you control have protection from black.
			Black creatures you control have protection from white.
		`,
		},
	}
}
