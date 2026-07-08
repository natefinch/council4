package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AetherChaser is the card definition for Aether Chaser.
//
// Type: Creature — Human Artificer
// Cost: {1}{R}
//
// Oracle text:
//
//	First strike
//	When this creature enters, you get {E}{E} (two energy counters).
//	Whenever this creature attacks, you may pay {E}{E}. If you do, create a 1/1 colorless Servo artifact creature token.
var AetherChaser = newAetherChaser

func newAetherChaser() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Aether Chaser",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Artificer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
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
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(aetherChaserToken),
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
			First strike
			When this creature enters, you get {E}{E} (two energy counters).
			Whenever this creature attacks, you may pay {E}{E}. If you do, create a 1/1 colorless Servo artifact creature token.
		`,
		},
	}
}

var aetherChaserToken = newAetherChaserToken()

func newAetherChaserToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Servo",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Servo},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
