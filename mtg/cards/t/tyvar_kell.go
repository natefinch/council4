package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TyvarKell is the card definition for Tyvar Kell.
//
// Type: Legendary Planeswalker — Tyvar
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Elves you control have "{T}: Add {B}."
//	+1: Put a +1/+1 counter on up to one target Elf. Untap it. It gains deathtouch until end of turn.
//	0: Create a 1/1 green Elf Warrior creature token.
//	−6: You get an emblem with "Whenever you cast an Elf spell, it gains haste until end of turn and you draw two cards."
var TyvarKell = newTyvarKell

func newTyvarKell() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Tyvar Kell",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Tyvar},
			Loyalty:    opt.Val(3),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Elves you control have \"{T}: Add {B}.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.TapManaAbility(mana.B)),
							},
						},
					},
				},
			},
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+1: Put a +1/+1 counter on up to one target Elf. Untap it. It gains deathtouch until end of turn.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 1,
										Constraint: "up to one target Elf",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Elf}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(0),
										},
									},
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Deathtouch},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
					Text: "0: Create a 1/1 green Elf Warrior creature token.",
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(tyvarKellToken),
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
					Text:        "−6: You get an emblem with \"Whenever you cast an Elf spell, it gains haste until end of turn and you draw two cards.\"",
					LoyaltyCost: -6,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.TriggeredAbility{
													Text: "Whenever you cast an Elf spell, it gains haste until end of turn and you draw two cards.",
													Trigger: game.TriggerCondition{
														Type: game.TriggerWhenever,
														Pattern: game.TriggerPattern{
															Event:           game.EventSpellCast,
															Controller:      game.TriggerControllerYou,
															CardSelection:   game.Selection{SubtypesAny: []types.Sub{types.Elf}},
															StackObjectKind: game.StackSpell,
															CounterKind:     counter.PlusOnePlusOne,
														},
													},
													Content: game.AbilityContent{
														Modes: []game.Mode{
															game.Mode{
																Sequence: []game.Instruction{
																	game.Instruction{
																		Primitive: game.ApplyContinuous{
																			Object: opt.Val(game.EventStackObjectReference()),
																			ContinuousEffects: []game.ContinuousEffect{
																				game.ContinuousEffect{
																					Duration:    game.DurationPermanent,
																					Layer:       game.LayerAbility,
																					AddKeywords: []game.Keyword{game.Haste},
																				},
																			},
																			Duration: game.DurationUntilEndOfTurn,
																		},
																	},
																	game.Instruction{
																		Primitive: game.Draw{
																			Amount: game.Fixed(2),
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
			Elves you control have "{T}: Add {B}."
			+1: Put a +1/+1 counter on up to one target Elf. Untap it. It gains deathtouch until end of turn.
			0: Create a 1/1 green Elf Warrior creature token.
			−6: You get an emblem with "Whenever you cast an Elf spell, it gains haste until end of turn and you draw two cards."
		`,
		},
	}
}

var tyvarKellToken = newTyvarKellToken()

func newTyvarKellToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elf Warrior",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
