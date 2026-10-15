package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Burrowing is the card definition for Burrowing.
//
// Type: Enchantment — Aura
// Cost: {R}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has mountainwalk. (It can't be blocked as long as defending player controls a Mountain.)
var Burrowing = newBurrowing

func newBurrowing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Burrowing",
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
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature has mountainwalk. (It can't be blocked as long as defending player controls a Mountain.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.StaticAbility{
									Text: "Mountainwalk",
									KeywordAbilities: []game.KeywordAbility{
										game.LandwalkKeyword{
											Subtype: types.Mountain,
										},
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has mountainwalk. (It can't be blocked as long as defending player controls a Mountain.)
		`,
		},
	}
}
