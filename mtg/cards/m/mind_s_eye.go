package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MindSEye is the card definition for Mind's Eye.
//
// Type: Artifact
// Cost: {5}
//
// Oracle text:
//
//	Whenever an opponent draws a card, you may pay {1}. If you do, draw a card.
var MindSEye = newMindSEye

func newMindSEye() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mind's Eye",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever an opponent draws a card, you may pay {1}. If you do, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventCardDrawn,
							Player:          game.TriggerPlayerOpponent,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1}?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			Whenever an opponent draws a card, you may pay {1}. If you do, draw a card.
		`,
		},
	}
}
