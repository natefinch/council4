package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CoercedToKill is the card definition for Coerced to Kill.
//
// Type: Enchantment — Aura
// Cost: {3}{U}{B}
//
// Oracle text:
//
//	Enchant creature
//	You control enchanted creature.
//	Enchanted creature has base power and toughness 1/1, has deathtouch, and is an Assassin in addition to its other types.
var CoercedToKill = newCoercedToKill

func newCoercedToKill() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Coerced to Kill",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.B,
			}),
			Colors:   []color.Color{color.Black, color.Blue},
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
					Text: "You control enchanted creature.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:      game.DurationPermanent,
							Group:         game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:         game.LayerControl,
							NewController: opt.Val(game.Player1),
						},
					},
				},
				game.StaticAbility{
					Text: "Enchanted creature has base power and toughness 1/1, has deathtouch, and is an Assassin in addition to its other types.",
					ContinuousEffects: []game.ContinuousEffect{
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
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Deathtouch},
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerType,
							AddSubtypes: []types.Sub{types.Assassin},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			You control enchanted creature.
			Enchanted creature has base power and toughness 1/1, has deathtouch, and is an Assassin in addition to its other types.
		`,
		},
	}
}
