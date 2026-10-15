package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CultivatorColossus is the card definition for Cultivator Colossus.
//
// Type: Creature — Plant Beast
// Cost: {4}{G}{G}{G}
//
// Oracle text:
//
//	Trample
//	Cultivator Colossus's power and toughness are each equal to the number of lands you control.
//	When this creature enters, you may put a land card from your hand onto the battlefield tapped. If you do, draw a card and repeat this process.
var CultivatorColossus = newCultivatorColossus

func newCultivatorColossus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cultivator Colossus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Plant, types.Beast},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, you may put a land card from your hand onto the battlefield tapped. If you do, draw a card and repeat this process.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.RepeatProcess{
											Body: game.AbilityContent{
												Modes: []game.Mode{
													game.Mode{
														Sequence: []game.Instruction{
															game.Instruction{
																Primitive: game.ChooseFromZone{
																	Player:     game.ControllerReference(),
																	SourceZone: zone.Hand,
																	Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}},
																	Quantity:   game.Fixed(1),
																	Destination: game.ChooseDestination{
																		Zone: zone.Battlefield,
																	},
																	Riders: game.ChooseRiders{
																		EntersTapped: true,
																	},
																	Prompt: "Choose a card to put onto the battlefield",
																},
																Optional:      true,
																PublishResult: game.ResultKey("if-you-do"),
																LocalProducts: game.LocalProducts{
																	Results: []game.ResultKey{game.ResultKey("if-you-do")},
																},
															},
															game.Instruction{
																Primitive: game.Draw{
																	Amount: game.Fixed(1),
																	Player: game.ControllerReference(),
																},
																ResultGate: opt.Val(game.InstructionResultGate{
																	Key:       game.ResultKey("if-you-do"),
																	Succeeded: game.TriTrue,
																}),
																PublishResult: game.ResultKey("repeat-process-continue"),
															},
														},
													},
												},
												MinModes: 1,
												MaxModes: 1,
											},
											ContinueResult: game.ResultKey("repeat-process-continue"),
										},
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
			Trample
			Cultivator Colossus's power and toughness are each equal to the number of lands you control.
			When this creature enters, you may put a land card from your hand onto the battlefield tapped. If you do, draw a card and repeat this process.
		`,
		},
	}
}
