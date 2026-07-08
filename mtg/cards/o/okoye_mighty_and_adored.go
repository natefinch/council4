package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OkoyeMightyAndAdored is the card definition for Okoye, Mighty and Adored.
//
// Type: Legendary Creature — Human Warrior Hero
// Cost: {2}{G}{W}
//
// Oracle text:
//
//	When Okoye enters, you become the monarch.
//	At the beginning of combat on your turn, put a +1/+1 counter on target creature. Whenever that creature attacks the monarch this turn, it gains double strike and trample until end of turn.
var OkoyeMightyAndAdored = newOkoyeMightyAndAdored

func newOkoyeMightyAndAdored() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Okoye, Mighty and Adored",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior, types.Hero},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepBeginningOfCombat,
						},
					},
					Content: game.Mode{
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
							{
								Primitive: game.AddCounter{
									Amount:        game.Fixed(1),
									Object:        game.TargetPermanentReference(0),
									CounterKind:   counter.PlusOnePlusOne,
									PublishLinked: game.LinkedKey("delayed-target-1"),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										EventPattern: opt.Val(game.TriggerPattern{
											Event:            game.EventAttackerDeclared,
											Player:           game.TriggerPlayerMonarch,
											AttackerCaptured: true,
											AttackRecipient:  game.AttackRecipientPlayer,
										}),
										Window:                 game.DelayedWindowThisTurn,
										CapturedAttackerObject: opt.Val(game.LinkedObjectReference("delayed-target-1")),
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.ApplyContinuous{
														Object: opt.Val(game.EventPermanentReference()),
														ContinuousEffects: []game.ContinuousEffect{
															game.ContinuousEffect{
																Layer: game.LayerAbility,
																AddKeywords: []game.Keyword{
																	game.DoubleStrike,
																	game.Trample,
																},
															},
														},
														Duration: game.DurationUntilEndOfTurn,
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When Okoye enters, you become the monarch.
			At the beginning of combat on your turn, put a +1/+1 counter on target creature. Whenever that creature attacks the monarch this turn, it gains double strike and trample until end of turn.
		`,
		},
	}
}
