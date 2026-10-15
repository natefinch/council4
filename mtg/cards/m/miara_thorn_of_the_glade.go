package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MiaraThornOfTheGlade is the card definition for Miara, Thorn of the Glade.
//
// Type: Legendary Creature — Elf Scout
// Cost: {1}{B}
//
// Oracle text:
//
//	Whenever Miara or another Elf you control dies, you may pay {1} and 1 life. If you do, draw a card.
//	Partner (You can have two commanders if both have partner.)
var MiaraThornOfTheGlade = newMiaraThornOfTheGlade

func newMiaraThornOfTheGlade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Miara, Thorn of the Glade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Scout},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Miara or another Elf you control dies, you may pay {1} and 1 life. If you do, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentDied,
							Controller:             game.TriggerControllerYou,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Elf}},
							SubjectSelectionOrSelf: true,
							StackObjectKind:        game.StackSpell,
							CounterKind:            counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1} and 1 life?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
												}),
												AdditionalCosts: []cost.Additional{
													cost.Additional{
														Kind:        cost.AdditionalPayLife,
														Text:        "1 life",
														Amount:      1,
														CounterKind: counter.PlusOnePlusOne,
													},
												},
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
			Whenever Miara or another Elf you control dies, you may pay {1} and 1 life. If you do, draw a card.
			Partner (You can have two commanders if both have partner.)
		`,
		},
	}
}
