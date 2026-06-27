package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThrivingTurtle is the card definition for Thriving Turtle.
//
// Type: Creature — Turtle
// Cost: {U}
//
// Oracle text:
//
//	When this creature enters, you get {E}{E} (two energy counters).
//	Whenever this creature attacks, you may pay {E}{E}. If you do, put a +1/+1 counter on it.
var ThrivingTurtle = newThrivingTurtle()

func newThrivingTurtle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Thriving Turtle",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Turtle},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
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
								Primitive: game.AddPlayerCounter{
									Amount:      game.Fixed(2),
									Player:      game.ControllerReference(),
									CounterKind: counter.Energy,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {E}{E}?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:   cost.AdditionalEnergy,
												Text:   "pay {E}{E}",
												Amount: 2,
											},
										},
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.EventPermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
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
			When this creature enters, you get {E}{E} (two energy counters).
			Whenever this creature attacks, you may pay {E}{E}. If you do, put a +1/+1 counter on it.
		`,
		},
	}
}
