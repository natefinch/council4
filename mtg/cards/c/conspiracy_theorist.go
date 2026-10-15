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

// ConspiracyTheorist is the card definition for Conspiracy Theorist.
//
// Type: Creature — Human Shaman
// Cost: {1}{R}
//
// Oracle text:
//
//	Whenever this creature attacks, you may pay {1} and discard a card. If you do, draw a card.
//	Whenever you discard one or more nonland cards, you may exile one of them from your graveyard. If you do, you may cast it this turn.
var ConspiracyTheorist = newConspiracyTheorist

func newConspiracyTheorist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Conspiracy Theorist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, you may pay {1} and discard a card. If you do, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
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
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1} and discard a card?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
												}),
												AdditionalCosts: []cost.Additional{
													cost.Additional{
														Kind:        cost.AdditionalDiscard,
														Text:        "discard a card",
														Amount:      1,
														Source:      zone.Hand,
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
				game.TriggeredAbility{
					Text: "Whenever you discard one or more nonland cards, you may exile one of them from your graveyard. If you do, you may cast it this turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventCardDiscarded,
							Player:          game.TriggerPlayerYou,
							CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Land}},
							StackObjectKind: game.StackSpell,
							OneOrMore:       true,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ExileForPlay{
											Card: game.CardReference{
												Kind: game.CardReferenceEvent,
											},
											FromZone:        zone.Graveyard,
											Duration:        game.DurationThisTurn,
											Cast:            true,
											SelectFromBatch: true,
										},
										Optional: true,
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
			Whenever this creature attacks, you may pay {1} and discard a card. If you do, draw a card.
			Whenever you discard one or more nonland cards, you may exile one of them from your graveyard. If you do, you may cast it this turn.
		`,
		},
	}
}
