package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SinisterStrength is the card definition for Sinister Strength.
//
// Type: Enchantment — Aura
// Cost: {1}{B}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +3/+1 and is black.
var SinisterStrength = newSinisterStrength

func newSinisterStrength() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sinister Strength",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
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
					Text: "Enchanted creature gets +3/+1 and is black.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     3,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:  game.DurationPermanent,
							Group:     game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:     game.LayerColor,
							SetColors: []color.Color{color.Black},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +3/+1 and is black.
		`,
		},
	}
}
