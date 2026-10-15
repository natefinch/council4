package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RipplesOfUndeath is the card definition for Ripples of Undeath.
//
// Type: Enchantment
// Cost: {1}{B}
//
// Oracle text:
//
//	At the beginning of your first main phase, mill three cards. Then you may pay {1} and 3 life. If you do, put a card from among those cards into your hand.
var RipplesOfUndeath = newRipplesOfUndeath

func newRipplesOfUndeath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ripples of Undeath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your first main phase, mill three cards. Then you may pay {1} and 3 life. If you do, put a card from among those cards into your hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepPrecombatMain,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:        game.Fixed(3),
											Player:        game.ControllerReference(),
											Destination:   zone.Graveyard,
											PublishLinked: game.LinkedKey("milled-cards"),
										},
									},
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1} and 3 life?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
												}),
												AdditionalCosts: []cost.Additional{
													cost.Additional{
														Kind:        cost.AdditionalPayLife,
														Text:        "3 life",
														Amount:      3,
														CounterKind: counter.PlusOnePlusOne,
													},
												},
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Graveyard,
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Hand,
											},
											Riders: game.ChooseRiders{
												FromLinked: game.LinkedKey("milled-cards"),
											},
											Prompt: "Choose a card to return to your hand",
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
			At the beginning of your first main phase, mill three cards. Then you may pay {1} and 3 life. If you do, put a card from among those cards into your hand.
		`,
		},
	}
}
