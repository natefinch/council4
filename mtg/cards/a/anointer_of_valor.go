package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AnointerOfValor is the card definition for Anointer of Valor.
//
// Type: Creature — Angel
// Cost: {5}{W}
//
// Oracle text:
//
//	Flying
//	Whenever a creature attacks, you may pay {3}. When you do, put a +1/+1 counter on that creature.
var AnointerOfValor = newAnointerOfValor

func newAnointerOfValor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Anointer of Valor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {3}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(3),
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.CreateReflexiveTrigger{
									Trigger: game.ReflexiveTriggerDef{
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.AddCounter{
														Amount:      game.Fixed(1),
														Object:      game.EventPermanentReference(),
														CounterKind: counter.PlusOnePlusOne,
													},
												},
											},
										}.Ability(),
									},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever a creature attacks, you may pay {3}. When you do, put a +1/+1 counter on that creature.
		`,
		},
	}
}
