package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChamberOfManipulation is the card definition for Chamber of Manipulation.
//
// Type: Enchantment — Aura
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Enchant land
//	Enchanted land has "{T}, Discard a card: Gain control of target creature until end of turn."
var ChamberOfManipulation = newChamberOfManipulation

func newChamberOfManipulation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Chamber of Manipulation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
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
					Text: "Enchanted land has \"{T}, Discard a card: Gain control of target creature until end of turn.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{T}, Discard a card: Gain control of target creature until end of turn.",
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:        cost.AdditionalTap,
											CounterKind: counter.PlusOnePlusOne,
										},
										cost.Additional{
											Kind:        cost.AdditionalDiscard,
											Text:        "Discard a card",
											Amount:      1,
											Source:      zone.Hand,
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
																	Duration:      game.DurationPermanent,
																	Layer:         game.LayerControl,
																	NewController: opt.Val(game.Player1),
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
			OracleText: `
			Enchant land
			Enchanted land has "{T}, Discard a card: Gain control of target creature until end of turn."
		`,
		},
	}
}
