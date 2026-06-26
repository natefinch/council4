package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SoulNet is the card definition for Soul Net.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	Whenever a creature dies, you may pay {1}. If you do, you gain 1 life.
var SoulNet = newSoulNet()

func newSoulNet() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Soul Net",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
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
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
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
			Whenever a creature dies, you may pay {1}. If you do, you gain 1 life.
		`,
		},
	}
}
