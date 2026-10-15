package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LilianaSDevotee is the card definition for Liliana's Devotee.
//
// Type: Creature — Human Warlock
// Cost: {2}{B}
//
// Oracle text:
//
//	Zombies you control get +1/+0.
//	At the beginning of your end step, if a creature died this turn, you may pay {1}{B}. If you do, create a 2/2 black Zombie creature token.
var LilianaSDevotee = newLilianaSDevotee

func newLilianaSDevotee() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Liliana's Devotee",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warlock},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Zombies you control get +1/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Zombie")}}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your end step, if a creature died this turn, you may pay {1}{B}. If you do, create a 2/2 black Zombie creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepEnd,
						},
						InterveningIf: "if a creature died this turn",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if a creature died this turn",
							SourceCounterKind: counter.PlusOnePlusOne,
							EventHistory: opt.Val(game.EventHistoryCondition{
								Pattern: game.TriggerPattern{
									Event:            game.EventPermanentDied,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
									StackObjectKind:  game.StackSpell,
									CounterKind:      counter.PlusOnePlusOne,
								},
								Window: game.EventHistoryCurrentTurn,
							}),
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1}{B}?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
													cost.B,
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(lilianaSDevoteeToken),
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
			Zombies you control get +1/+0.
			At the beginning of your end step, if a creature died this turn, you may pay {1}{B}. If you do, create a 2/2 black Zombie creature token.
		`,
		},
	}
}

var lilianaSDevoteeToken = newLilianaSDevoteeToken()

func newLilianaSDevoteeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
