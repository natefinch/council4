package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CloutOfTheDominus is the card definition for Clout of the Dominus.
//
// Type: Enchantment — Aura
// Cost: {U/R}
//
// Oracle text:
//
//	Enchant creature
//	As long as enchanted creature is blue, it gets +1/+1 and has shroud. (It can't be the target of spells or abilities.)
//	As long as enchanted creature is red, it gets +1/+1 and has haste.
var CloutOfTheDominus = newCloutOfTheDominus()

func newCloutOfTheDominus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Clout of the Dominus",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.U, mana.R),
			}),
			Colors:   []color.Color{color.Red, color.Blue},
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
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Blue}}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Shroud,
							},
						},
					},
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Red}}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Haste,
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			As long as enchanted creature is blue, it gets +1/+1 and has shroud. (It can't be the target of spells or abilities.)
			As long as enchanted creature is red, it gets +1/+1 and has haste.
		`,
		},
	}
}
