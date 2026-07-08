package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TempestTechnique is the card definition for Tempest Technique.
//
// Type: Enchantment — Aura
// Cost: {3}{W}
//
// Oracle text:
//
//	Storm (When you cast this spell, copy it for each spell cast before it this turn. You may choose new targets for the copies. Copies become tokens.)
//	Enchant creature you control
//	Enchanted creature gets +1/+1 for each enchantment you control.
var TempestTechnique = newTempestTechnique

func newTempestTechnique() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Tempest Technique",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.StormStaticBody,
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature},
						Controller:       game.ControllerYou,
					}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerPowerToughnessModify,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou}),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou}),
							}),
						},
					},
				},
			},
			OracleText: `
			Storm (When you cast this spell, copy it for each spell cast before it this turn. You may choose new targets for the copies. Copies become tokens.)
			Enchant creature you control
			Enchanted creature gets +1/+1 for each enchantment you control.
		`,
		},
	}
}
