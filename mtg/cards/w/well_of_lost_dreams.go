package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WellOfLostDreams is the card definition for Well of Lost Dreams.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	Whenever you gain life, you may pay {X}, where X is less than or equal to the amount of life you gained. If you do, draw X cards.
var WellOfLostDreams = newWellOfLostDreams

func newWellOfLostDreams() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Well of Lost Dreams",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventLifeGained,
							Player: game.TriggerPlayerYou,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PayRepeatedly{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {1} to draw a card?",
										Payer:  opt.Val(game.ControllerReference()),
										ManaCost: opt.Val(cost.Mana{
											cost.O(1),
										}),
									},
									PublishCount: "variable-pay-scaled-draw-count",
									MaxCount: opt.Val(&game.DynamicAmount{
										Kind:       game.DynamicAmountEventLifeChange,
										Multiplier: 1,
									}),
								},
								PublishResult: game.ResultKey("variable-pay-scaled-draw-count"),
							},
							{
								Primitive: game.Draw{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:      game.DynamicAmountChosenNumber,
										ResultKey: game.ResultKey("variable-pay-scaled-draw-count"),
									}),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "variable-pay-scaled-draw-count",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you gain life, you may pay {X}, where X is less than or equal to the amount of life you gained. If you do, draw X cards.
		`,
		},
	}
}
