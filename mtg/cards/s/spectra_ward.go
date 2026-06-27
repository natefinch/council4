package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpectraWard is the card definition for Spectra Ward.
//
// Type: Enchantment — Aura
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +2/+2 and has protection from each color. This effect doesn't remove Auras. (It can't be blocked, targeted, or dealt damage by anything that's white, blue, black, red, or green.)
var SpectraWard = newSpectraWard()

func newSpectraWard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Spectra Ward",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
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
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ProtectionFromEachColorStaticAbility()),
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +2/+2 and has protection from each color. This effect doesn't remove Auras. (It can't be blocked, targeted, or dealt damage by anything that's white, blue, black, red, or green.)
		`,
		},
	}
}
