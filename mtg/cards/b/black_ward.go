package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlackWard is the card definition for Black Ward.
//
// Type: Enchantment — Aura
// Cost: {W}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has protection from black. This effect doesn't remove this Aura.
var BlackWard = newBlackWard

func newBlackWard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Black Ward",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ProtectionFromColorsStaticAbility(color.Black)),
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has protection from black. This effect doesn't remove this Aura.
		`,
		},
	}
}
