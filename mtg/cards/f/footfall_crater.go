package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FootfallCrater is the card definition for Footfall Crater.
//
// Type: Enchantment — Aura
// Cost: {R}
//
// Oracle text:
//
//	Enchant land
//	Enchanted land has "{T}: Target creature gains trample and haste until end of turn."
//	Cycling {1} ({1}, Discard this card: Draw a card.)
var FootfallCrater = newFootfallCrater

func newFootfallCrater() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Footfall Crater",
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
					Constraint: "land",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
				}),
				game.StaticAbility{
					Text: "Enchanted land has \"{T}: Target creature gains trample and haste until end of turn.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{T}: Target creature gains trample and haste until end of turn.",
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
												Targets: []game.TargetSpec{
													game.TargetSpec{
														MinTargets: 1,
														MaxTargets: 1,
														Constraint: "target creature",
														Allow:      game.TargetAllowPermanent,
														Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.ApplyContinuous{
															Object: opt.Val(game.TargetPermanentReference(0)),
															ContinuousEffects: []game.ContinuousEffect{
																game.ContinuousEffect{
																	Duration:    game.DurationPermanent,
																	Layer:       game.LayerAbility,
																	AddKeywords: []game.Keyword{game.Trample, game.Haste},
																},
															},
															Duration: game.DurationUntilEndOfTurn,
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
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1)}),
			},
			OracleText: `
			Enchant land
			Enchanted land has "{T}: Target creature gains trample and haste until end of turn."
			Cycling {1} ({1}, Discard this card: Draw a card.)
		`,
		},
	}
}
