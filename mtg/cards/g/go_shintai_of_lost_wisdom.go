package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GoShintaiOfLostWisdom is the card definition for Go-Shintai of Lost Wisdom.
//
// Type: Legendary Enchantment Creature — Shrine
// Cost: {1}{U}
//
// Oracle text:
//
//	Flying
//	At the beginning of your end step, you may pay {1}. When you do, target player mills X cards, where X is the number of Shrines you control. (To mill a card, a player puts the top card of their library into their graveyard.)
var GoShintaiOfLostWisdom = newGoShintaiOfLostWisdom

func newGoShintaiOfLostWisdom() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Go-Shintai of Lost Wisdom",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment, types.Creature},
			Subtypes:   []types.Sub{types.Shrine},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your end step, you may pay {1}. When you do, target player mills X cards, where X is the number of Shrines you control. (To mill a card, a player puts the top card of their library into their graveyard.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepEnd,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1}?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
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
																	Constraint: "target player",
																	Allow:      game.TargetAllowPlayer,
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.MoveTopOfLibrary{
																		Amount: game.Dynamic(game.DynamicAmount{
																			Kind:       game.DynamicAmountCountSelector,
																			Multiplier: 1,
																			Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Shrine")}, Controller: game.ControllerYou}),
																		}),
																		Player:      game.TargetPlayerReference(0),
																		Destination: zone.Graveyard,
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
			OracleText: `
			Flying
			At the beginning of your end step, you may pay {1}. When you do, target player mills X cards, where X is the number of Shrines you control. (To mill a card, a player puts the top card of their library into their graveyard.)
		`,
		},
	}
}
