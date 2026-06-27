package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThrivingIbex is the card definition for Thriving Ibex.
//
// Type: Creature — Goat
// Cost: {3}{W}
//
// Oracle text:
//
//	When this creature enters, you get {E}{E} (two energy counters).
//	Whenever this creature attacks, you may pay {E}{E}. If you do, put a +1/+1 counter on it.
var ThrivingIbex = newThrivingIbex()

func newThrivingIbex() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Thriving Ibex",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goat},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
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
