package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VenomousBrutalizer is the card definition for Venomous Brutalizer.
//
// Type: Creature — Phyrexian Knight
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Toxic 3 (Players dealt combat damage by this creature also get three poison counters.)
//	When this creature enters, you may pay {1}{G}. If you do, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
var VenomousBrutalizer = newVenomousBrutalizer

func newVenomousBrutalizer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Venomous Brutalizer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Knight},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Toxic 3",
					KeywordAbilities: []game.KeywordAbility{
						game.ToxicKeyword{
							Amount: 3,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, you may pay {1}{G}. If you do, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)",
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
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1}{G}?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
													cost.G,
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.Proliferate{
											Amount: game.Fixed(1),
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
			Toxic 3 (Players dealt combat damage by this creature also get three poison counters.)
			When this creature enters, you may pay {1}{G}. If you do, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
		`,
		},
	}
}
