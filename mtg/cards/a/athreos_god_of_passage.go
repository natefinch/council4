package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AthreosGodOfPassage is the card definition for Athreos, God of Passage.
//
// Type: Legendary Enchantment Creature — God
// Cost: {1}{W}{B}
//
// Oracle text:
//
//	Indestructible
//	As long as your devotion to white and black is less than seven, Athreos isn't a creature.
//	Whenever another creature you own dies, return it to your hand unless target opponent pays 3 life.
var AthreosGodOfPassage = newAthreosGodOfPassage

func newAthreosGodOfPassage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Athreos, God of Passage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment, types.Creature},
			Subtypes:   []types.Sub{types.God},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
				game.StaticAbility{
					Text: "As long as your devotion to white and black is less than seven, Athreos isn't a creature.",
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerDevotion,
								Op:        compare.LessThan,
								Value:     7,
								Colors:    []color.Color{color.White, color.Black},
							},
						},
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerType,
							RemoveTypes:    []types.Card{types.Creature},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever another creature you own dies, return it to your hand unless target opponent pays 3 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							ExcludeSelf:      true,
							Player:           game.TriggerPlayerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target opponent",
										Allow:      game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "3 life?",
												Payer:  opt.Val(game.TargetPlayerReference(0)),
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
										PublishResult: game.ResultKey("target-player-paid"),
									},
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceEvent,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("target-player-paid"),
											Succeeded: game.TriFalse,
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
			Indestructible
			As long as your devotion to white and black is less than seven, Athreos isn't a creature.
			Whenever another creature you own dies, return it to your hand unless target opponent pays 3 life.
		`,
		},
	}
}
