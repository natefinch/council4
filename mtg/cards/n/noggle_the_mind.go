package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NoggleTheMind is the card definition for Noggle the Mind.
//
// Type: Enchantment — Aura
// Cost: {1}{U}
//
// Oracle text:
//
//	Flash
//	Enchant creature
//	Enchanted creature loses all abilities and is a colorless Noggle with base power and toughness 1/1. (It loses all colors and all other creature types.)
var NoggleTheMind = newNoggleTheMind

func newNoggleTheMind() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Noggle the Mind",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature loses all abilities and is a colorless Noggle with base power and toughness 1/1. (It loses all colors and all other creature types.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:           game.DurationPermanent,
							Group:              game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:              game.LayerAbility,
							RemoveAllAbilities: true,
						},
						game.ContinuousEffect{
							Duration:     game.DurationPermanent,
							Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:        game.LayerColor,
							SetColorless: true,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerType,
							SetSubtypes: []types.Sub{types.Noggle},
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 1,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 1,
							}),
						},
					},
				},
			},
			OracleText: `
			Flash
			Enchant creature
			Enchanted creature loses all abilities and is a colorless Noggle with base power and toughness 1/1. (It loses all colors and all other creature types.)
		`,
		},
	}
}
