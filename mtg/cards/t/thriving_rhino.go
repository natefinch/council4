package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThrivingRhino is the card definition for Thriving Rhino.
//
// Type: Creature — Rhino
// Cost: {2}{G}
//
// Oracle text:
//
//	When this creature enters, you get {E}{E} (two energy counters).
//	Whenever this creature attacks, you may pay {E}{E}. If you do, put a +1/+1 counter on it.
var ThrivingRhino = newThrivingRhino()

func newThrivingRhino() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Thriving Rhino",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rhino},
			Power:     opt.Val(game.PT{Value: 2}),
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
