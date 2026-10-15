package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SlinzaTheSpikedStampede is the card definition for Slinza, the Spiked Stampede.
//
// Type: Legendary Creature — Beast
// Cost: {4}{G}
//
// Oracle text:
//
//	Beast spells you cast cost {2} less to cast.
//	Each other Beast creature you control enters with an additional +1/+1 counter on it.
//	Whenever Slinza or another creature with power 4 or greater enters, you may pay {1}{R/G}. When you do, Slinza fights target creature you don't control.
var SlinzaTheSpikedStampede = newSlinzaTheSpikedStampede

func newSlinzaTheSpikedStampede() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Slinza, the Spiked Stampede",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Beast},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Beast spells you cast cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Beast}},
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Slinza or another creature with power 4 or greater enters, you may pay {1}{R/G}. When you do, Slinza fights target creature you don't control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event: game.EventPermanentEnteredBattlefield,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{
								Op:    compare.GreaterOrEqual,
								Value: 4,
							})},
							SubjectSelectionOrSelf: true,
							StackObjectKind:        game.StackSpell,
							CounterKind:            counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1}{R/G}?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
													cost.HybridMana(mana.R, mana.G),
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.CreateReflexiveTrigger{
											Trigger: game.ReflexiveTriggerDef{
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Targets: []game.TargetSpec{
																game.TargetSpec{
																	MinTargets: 1,
																	MaxTargets: 1,
																	Constraint: "target creature you don't control",
																	Allow:      game.TargetAllowPermanent,
																	Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerNotYou}),
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.Fight{
																		Object:        game.SourcePermanentReference(),
																		RelatedObject: game.TargetPermanentReference(0),
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
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("controller-paid"),
											Succeeded: game.TriTrue,
										}),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersGroupReplacement("Each other Beast creature you control enters with an additional +1/+1 counter on it.", &game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Beast")}, Controller: game.ControllerYou, ExcludeSource: true}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Beast spells you cast cost {2} less to cast.
			Each other Beast creature you control enters with an additional +1/+1 counter on it.
			Whenever Slinza or another creature with power 4 or greater enters, you may pay {1}{R/G}. When you do, Slinza fights target creature you don't control.
		`,
		},
	}
}
