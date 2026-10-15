package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SquirrelNest is the card definition for Squirrel Nest.
//
// Type: Enchantment — Aura
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Enchant land
//	Enchanted land has "{T}: Create a 1/1 green Squirrel creature token."
var SquirrelNest = newSquirrelNest

func newSquirrelNest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Squirrel Nest",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "land",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
				}),
				game.StaticAbility{
					Text: "Enchanted land has \"{T}: Create a 1/1 green Squirrel creature token.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{T}: Create a 1/1 green Squirrel creature token.",
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:        cost.AdditionalTap,
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.CreateToken{
															Amount: game.Fixed(1),
															Source: game.TokenDef(squirrelNestToken),
														},
													},
												},
											},
										},
										MinModes: 1,
										MaxModes: 1,
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			Enchant land
			Enchanted land has "{T}: Create a 1/1 green Squirrel creature token."
		`,
		},
	}
}

var squirrelNestToken = newSquirrelNestToken()

func newSquirrelNestToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Squirrel",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Squirrel},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
