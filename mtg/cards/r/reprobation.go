package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Reprobation is the card definition for Reprobation.
//
// Type: Enchantment — Aura
// Cost: {1}{W}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature loses all abilities and is a Coward creature with base power and toughness 0/1. (It keeps all supertypes but loses all other types and creature types.)
var Reprobation = newReprobation

func newReprobation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Reprobation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
					Text: "Enchanted creature loses all abilities and is a Coward creature with base power and toughness 0/1. (It keeps all supertypes but loses all other types and creature types.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:           game.DurationPermanent,
							Group:              game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:              game.LayerAbility,
							RemoveAllAbilities: true,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerType,
							SetTypes:    []types.Card{types.Creature},
							SetSubtypes: []types.Sub{types.Coward},
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{}),
							SetToughness: opt.Val(game.PT{
								Value: 1,
							}),
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature loses all abilities and is a Coward creature with base power and toughness 0/1. (It keeps all supertypes but loses all other types and creature types.)
		`,
		},
	}
}
