package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Lignify is the card definition for Lignify.
//
// Type: Kindred Enchantment — Treefolk Aura
// Cost: {1}{G}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature is a Treefolk with base power and toughness 0/4 and loses all abilities.
var Lignify = newLignify

func newLignify() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lignify",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Kindred, types.Enchantment},
			Subtypes: []types.Sub{types.Treefolk, types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature is a Treefolk with base power and toughness 0/4 and loses all abilities.",
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
							SetSubtypes: []types.Sub{types.Treefolk},
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{}),
							SetToughness: opt.Val(game.PT{
								Value: 4,
							}),
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature is a Treefolk with base power and toughness 0/4 and loses all abilities.
		`,
		},
	}
}
