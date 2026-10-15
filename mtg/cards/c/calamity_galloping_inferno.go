package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CalamityGallopingInferno is the card definition for Calamity, Galloping Inferno.
//
// Type: Legendary Creature — Horse Mount
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Haste
//	Whenever Calamity attacks while saddled, choose a nonlegendary creature that saddled it this turn and create a tapped and attacking token that's a copy of it. Sacrifice that token at the beginning of the next end step. Repeat this process once.
//	Saddle 1
var CalamityGallopingInferno = newCalamityGallopingInferno

func newCalamityGallopingInferno() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Calamity, Galloping Inferno",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Horse, types.Mount},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.SaddleActivatedAbility(1),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Calamity attacks while saddled, choose a nonlegendary creature that saddled it this turn and create a tapped and attacking token that's a copy of it. Sacrifice that token at the beginning of the next end step. Repeat this process once.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:              game.EventAttackerDeclared,
							Source:             game.TriggerSourceSelf,
							StackObjectKind:    game.StackSpell,
							CounterKind:        counter.PlusOnePlusOne,
							AttackWhileSaddled: true,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.RepeatProcess{
											Times: game.Fixed(2),
											Body: game.AbilityContent{
												Modes: []game.Mode{
													game.Mode{
														Sequence: []game.Instruction{
															game.Instruction{
																Primitive: game.CreateToken{
																	Amount: game.Fixed(1),
																	Source: game.TokenCopyOf(game.TokenCopySpec{
																		Source: game.TokenCopySourceChosenFromGroup,
																		Group:  game.GroupRef(game.SaddleContributorsGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSupertype: types.Legendary})),
																	}),
																	EntryTapped:        true,
																	AttackSameAsSource: true,
																	PublishLinked:      game.LinkedKey("delayed-created-token-sacrifice-1"),
																},
																PublishResult: game.ResultKey("created-token-1"),
															},
															game.Instruction{
																Primitive: game.CreateDelayedTrigger{
																	Trigger: game.DelayedTriggerDef{
																		Timing: game.DelayedAtBeginningOfNextEndStep,
																		Content: game.AbilityContent{
																			Modes: []game.Mode{
																				game.Mode{
																					Sequence: []game.Instruction{
																						game.Instruction{
																							Primitive: game.Sacrifice{
																								Group: game.CapturedObjectsGroup(),
																							},
																						},
																					},
																				},
																			},
																			MinModes: 1,
																			MaxModes: 1,
																		},
																		CapturedObjectGroup: opt.Val(game.LinkedObjectReference("delayed-created-token-sacrifice-1")),
																	},
																},
																ResultGate: opt.Val(game.InstructionResultGate{
																	Key:       game.ResultKey("created-token-1"),
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
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Haste
			Whenever Calamity attacks while saddled, choose a nonlegendary creature that saddled it this turn and create a tapped and attacking token that's a copy of it. Sacrifice that token at the beginning of the next end step. Repeat this process once.
			Saddle 1
		`,
		},
	}
}
