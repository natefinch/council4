package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DownInTheValley is the card definition for Down in the Valley.
//
// Type: Enchantment — Saga
// Cost: {2}{G}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after IV.)
//	I — Search your library for a basic land card, reveal it, put it into your hand, then shuffle.
//	II — This Saga gains "Landfall — Whenever a land you control enters, create a 1/1 green Elf creature token."
//	III, IV — Elves you control get +1/+0 and gain vigilance until end of turn.
var DownInTheValley = newDownInTheValley

func newDownInTheValley() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Down in the Valley",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
					Chapters: []int{1},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:  zone.Library,
												Destination: zone.Hand,
												Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
												Reveal:      true,
											},
											Amount: game.Fixed(1),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "II — This Saga gains \"Landfall — Whenever a land you control enters, create a 1/1 green Elf creature token.\"",
					Chapters: []int{2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerAbility,
													AddAbilities: []game.Ability{
														&game.TriggeredAbility{
															Text: "Landfall — Whenever a land you control enters, create a 1/1 green Elf creature token.",
															Trigger: game.TriggerCondition{
																Type: game.TriggerWhenever,
																Pattern: game.TriggerPattern{
																	Event:            game.EventPermanentEnteredBattlefield,
																	Controller:       game.TriggerControllerYou,
																	SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
																	StackObjectKind:  game.StackSpell,
																	CounterKind:      counter.PlusOnePlusOne,
																},
															},
															Content: game.AbilityContent{
																Modes: []game.Mode{
																	game.Mode{
																		Sequence: []game.Instruction{
																			game.Instruction{
																				Primitive: game.CreateToken{
																					Amount: game.Fixed(1),
																					Source: game.TokenDef(downInTheValleyToken),
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
											Duration: game.DurationPermanent,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "III, IV — Elves you control get +1/+0 and gain vigilance until end of turn.",
					Chapters: []int{3, 4},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:   game.DurationPermanent,
													Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}, Controller: game.ControllerYou}),
													Layer:      game.LayerPowerToughnessModify,
													PowerDelta: 1,
												},
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}, Controller: game.ControllerYou}),
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Vigilance},
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
			},
			OracleText: `
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after IV.)
			I — Search your library for a basic land card, reveal it, put it into your hand, then shuffle.
			II — This Saga gains "Landfall — Whenever a land you control enters, create a 1/1 green Elf creature token."
			III, IV — Elves you control get +1/+0 and gain vigilance until end of turn.
		`,
		},
	}
}

var downInTheValleyToken = newDownInTheValleyToken()

func newDownInTheValleyToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elf",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
