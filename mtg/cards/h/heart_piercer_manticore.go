package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HeartPiercerManticore is the card definition for Heart-Piercer Manticore.
//
// Type: Creature — Manticore
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	When this creature enters, you may sacrifice another creature. When you do, this creature deals damage equal to that creature's power to any target.
//	Embalm {5}{R} ({5}{R}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Manticore with no mana cost. Embalm only as a sorcery.)
var HeartPiercerManticore = newHeartPiercerManticore

func newHeartPiercerManticore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Heart-Piercer Manticore",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Manticore},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.EmbalmActivatedBody(cost.Mana{cost.O(5), cost.R}, types.Sub("Manticore")),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, you may sacrifice another creature. When you do, this creature deals damage equal to that creature's power to any target.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SacrificePermanents{
											Player:               game.ControllerReference(),
											Amount:               game.Fixed(1),
											Selection:            game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true},
											PublishLinked:        game.LinkedKey("sacrificed-creature"),
											PublishObjectBinding: true,
										},
										Optional:      true,
										PublishResult: game.ResultKey("if-you-do"),
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
																	Constraint: "any target",
																	Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.Damage{
																		Amount: game.Dynamic(game.DynamicAmount{
																			Kind:       game.DynamicAmountObjectPower,
																			Multiplier: 1,
																			Object:     game.LinkedObjectReference("sacrificed-creature"),
																		}),
																		Recipient:    game.AnyTargetDamageRecipient(0),
																		DamageSource: opt.Val(game.SourcePermanentReference()),
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
											Key:       game.ResultKey("if-you-do"),
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
			When this creature enters, you may sacrifice another creature. When you do, this creature deals damage equal to that creature's power to any target.
			Embalm {5}{R} ({5}{R}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Manticore with no mana cost. Embalm only as a sorcery.)
		`,
		},
	}
}
