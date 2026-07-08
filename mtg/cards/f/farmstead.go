package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Farmstead is the card definition for Farmstead.
//
// Type: Enchantment — Aura
// Cost: {W}{W}{W}
//
// Oracle text:
//
//	Enchant land
//	Enchanted land has "At the beginning of your upkeep, you may pay {W}{W}. If you do, you gain 1 life."
var Farmstead = newFarmstead

func newFarmstead() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Farmstead",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:   []color.Color{color.White},
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
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerAt,
										Pattern: game.TriggerPattern{
											Event:      game.EventBeginningOfStep,
											Controller: game.TriggerControllerYou,
											Step:       game.StepUpkeep,
										},
									},
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.Pay{
													Payment: game.ResolutionPayment{
														Prompt: "Pay {W}{W}?",
														ManaCost: opt.Val(cost.Mana{
															cost.W,
															cost.W,
														}),
													},
												},
												PublishResult: game.ResultKey("controller-paid"),
											},
											{
												Primitive: game.GainLife{
													Amount: game.Fixed(1),
													Player: game.ControllerReference(),
												},
												ResultGate: opt.Val(game.InstructionResultGate{
													Key:       "controller-paid",
													Succeeded: game.TriTrue,
												}),
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			OracleText: `
			Enchant land
			Enchanted land has "At the beginning of your upkeep, you may pay {W}{W}. If you do, you gain 1 life."
		`,
		},
	}
}
