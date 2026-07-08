package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmblemOfTheWarmind is the card definition for Emblem of the Warmind.
//
// Type: Enchantment — Aura
// Cost: {1}{R}
//
// Oracle text:
//
//	Enchant creature you control
//	Creatures you control have haste.
var EmblemOfTheWarmind = newEmblemOfTheWarmind

func newEmblemOfTheWarmind() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Emblem of the Warmind",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
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
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							AddKeywords: []game.Keyword{
								game.Haste,
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature you control
			Creatures you control have haste.
		`,
		},
	}
}
