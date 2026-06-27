package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShieldOfTheOversoul is the card definition for Shield of the Oversoul.
//
// Type: Enchantment — Aura
// Cost: {2}{G/W}
//
// Oracle text:
//
//	Enchant creature
//	As long as enchanted creature is green, it gets +1/+1 and has indestructible. (Damage and effects that say "destroy" don't destroy it. If its toughness is 0 or less, it still dies.)
//	As long as enchanted creature is white, it gets +1/+1 and has flying.
var ShieldOfTheOversoul = newShieldOfTheOversoul()

func newShieldOfTheOversoul() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Shield of the Oversoul",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors:   []color.Color{color.Green, color.White},
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
						ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Green}}),
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
								game.Indestructible,
							},
						},
					},
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.White}}),
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
								game.Flying,
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			As long as enchanted creature is green, it gets +1/+1 and has indestructible. (Damage and effects that say "destroy" don't destroy it. If its toughness is 0 or less, it still dies.)
			As long as enchanted creature is white, it gets +1/+1 and has flying.
		`,
		},
	}
}
