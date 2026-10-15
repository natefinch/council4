package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JeweledTorque is the card definition for Jeweled Torque.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	As this artifact enters, choose a color.
//	Whenever a player casts a spell of the chosen color, you may pay {2}. If you do, you gain 2 life.
var JeweledTorque = newJeweledTorque

func newJeweledTorque() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Jeweled Torque",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a player casts a spell of the chosen color, you may pay {2}. If you do, you gain 2 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							CardSelection:   game.Selection{ColorChoice: game.ColorChoiceSourceEntry},
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
												Prompt: "Pay {2}?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(2),
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this artifact enters, choose a color."),
			},
			OracleText: `
			As this artifact enters, choose a color.
			Whenever a player casts a spell of the chosen color, you may pay {2}. If you do, you gain 2 life.
		`,
		},
	}
}
