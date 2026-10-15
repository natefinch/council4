package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZiatoraTheIncinerator is the card definition for Ziatora, the Incinerator.
//
// Type: Legendary Creature — Demon Dragon
// Cost: {3}{B}{R}{G}
//
// Oracle text:
//
//	Flying
//	At the beginning of your end step, you may sacrifice another creature. When you do, Ziatora deals damage equal to that creature's power to any target and you create three Treasure tokens.
var ZiatoraTheIncinerator = newZiatoraTheIncinerator

func newZiatoraTheIncinerator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Ziatora, the Incinerator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Demon, types.Dragon},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your end step, you may sacrifice another creature. When you do, Ziatora deals damage equal to that creature's power to any target and you create three Treasure tokens.",
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
																game.Instruction{
																	Primitive: game.CreateToken{
																		Amount: game.Fixed(3),
																		Source: game.TokenDef(ziatoraTheIncineratorToken),
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
			Flying
			At the beginning of your end step, you may sacrifice another creature. When you do, Ziatora deals damage equal to that creature's power to any target and you create three Treasure tokens.
		`,
		},
	}
}

var ziatoraTheIncineratorToken = newZiatoraTheIncineratorToken()

func newZiatoraTheIncineratorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Treasure",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Treasure},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}, Sacrifice this artifact: Add one mana of any color.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Choose{
											Choice: game.ResolutionChoice{
												Kind:   game.ResolutionChoiceMana,
												Prompt: "Choose a color",
												Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
											},
											PublishChoice: game.ChoiceKey("oracle-mana-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(1),
											ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
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
	}
}
