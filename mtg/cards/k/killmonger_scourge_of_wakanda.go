package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KillmongerScourgeOfWakanda is the card definition for Killmonger, Scourge of Wakanda.
//
// Type: Legendary Creature — Human Mercenary Villain
// Cost: {2}{B}{G}
//
// Oracle text:
//
//	When Killmonger enters, you may sacrifice another creature. When you do, destroy target nonland permanent an opponent controls.
//	As long as there are two or more creature cards in your graveyard, Killmonger gets +2/+1.
var KillmongerScourgeOfWakanda = newKillmongerScourgeOfWakanda

func newKillmongerScourgeOfWakanda() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Killmonger, Scourge of Wakanda",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Mercenary, types.Villain},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as there are two or more creature cards in your graveyard, Killmonger gets +2/+1.",
					Condition: opt.Val(game.Condition{
						Text:              "As long as there are two or more creature cards in your graveyard",
						SourceCounterKind: counter.PlusOnePlusOne,
						ControllerGraveyardCardOfTypeCountAtLeast: 2,
						ControllerGraveyardCountCardType:          types.Creature,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Killmonger enters, you may sacrifice another creature. When you do, destroy target nonland permanent an opponent controls.",
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
											Player:    game.ControllerReference(),
											Amount:    game.Fixed(1),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true},
										},
										Optional:      true,
										PublishResult: game.ResultKey("if-you-do"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("if-you-do")},
										},
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
																	Constraint: "target nonland permanent an opponent controls",
																	Allow:      game.TargetAllowPermanent,
																	Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerOpponent}),
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.Destroy{
																		Object: game.TargetPermanentReference(0),
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
			When Killmonger enters, you may sacrifice another creature. When you do, destroy target nonland permanent an opponent controls.
			As long as there are two or more creature cards in your graveyard, Killmonger gets +2/+1.
		`,
		},
	}
}
