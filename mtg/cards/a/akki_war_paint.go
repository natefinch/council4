package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AkkiWarPaint is the card definition for Akki War Paint.
//
// Type: Enchantment — Aura
// Cost: {R}
//
// Oracle text:
//
//	Enchant artifact or creature
//	As long as enchanted permanent is a creature, it gets +2/+1.
var AkkiWarPaint = newAkkiWarPaint()

func newAkkiWarPaint() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Akki War Paint",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "artifact or creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
				}),
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     2,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Enchant artifact or creature
			As long as enchanted permanent is a creature, it gets +2/+1.
		`,
		},
	}
}
