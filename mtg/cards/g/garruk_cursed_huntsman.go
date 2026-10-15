package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GarrukCursedHuntsman is the card definition for Garruk, Cursed Huntsman.
//
// Type: Legendary Planeswalker — Garruk
// Cost: {4}{B}{G}
//
// Oracle text:
//
//	0: Create two 2/2 black and green Wolf creature tokens with "When this token dies, put a loyalty counter on each Garruk you control."
//	−3: Destroy target creature. Draw a card.
//	−6: You get an emblem with "Creatures you control get +3/+3 and have trample."
var GarrukCursedHuntsman = newGarrukCursedHuntsman

func newGarrukCursedHuntsman() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Garruk, Cursed Huntsman",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Garruk},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text: "0: Create two 2/2 black and green Wolf creature tokens with \"When this token dies, put a loyalty counter on each Garruk you control.\"",
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(2),
											Source: game.TokenDef(garrukCursedHuntsmanToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−3: Destroy target creature. Draw a card.",
					LoyaltyCost: -3,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−6: You get an emblem with \"Creatures you control get +3/+3 and have trample.\"",
					LoyaltyCost: -6,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.StaticAbility{
													Text: "Creatures you control get +3/+3 and have trample.",
													ContinuousEffects: []game.ContinuousEffect{
														game.ContinuousEffect{
															Duration:       game.DurationPermanent,
															Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
															Layer:          game.LayerPowerToughnessModify,
															PowerDelta:     3,
															ToughnessDelta: 3,
														},
														game.ContinuousEffect{
															Duration:    game.DurationPermanent,
															Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
															Layer:       game.LayerAbility,
															AddKeywords: []game.Keyword{game.Trample},
														},
													},
												},
											},
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
			0: Create two 2/2 black and green Wolf creature tokens with "When this token dies, put a loyalty counter on each Garruk you control."
			−3: Destroy target creature. Draw a card.
			−6: You get an emblem with "Creatures you control get +3/+3 and have trample."
		`,
		},
	}
}

var garrukCursedHuntsmanToken = newGarrukCursedHuntsmanToken()

func newGarrukCursedHuntsmanToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wolf",
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this token dies, put a loyalty counter on each Garruk you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Garruk")}, Controller: game.ControllerYou}),
											CounterKind: counter.Loyalty,
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
		},
	}
}
