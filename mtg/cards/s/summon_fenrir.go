package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SummonFenrir is the card definition for Summon: Fenrir.
//
// Type: Enchantment Creature — Saga Wolf
// Cost: {2}{G}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I — Crescent Fang — Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
//	II — Heavenward Howl — When you next cast a creature spell this turn, that creature enters with an additional +1/+1 counter on it.
//	III — Ecliptic Growl — Draw a card if you control the creature with the greatest power or tied for the greatest power.
var SummonFenrir = newSummonFenrir

func newSummonFenrir() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Summon: Fenrir",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Saga, types.Wolf},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — Crescent Fang — Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
					Chapters: []int{1},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:   zone.Library,
												Destination:  zone.Battlefield,
												Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
												EntersTapped: true,
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
					Text:     "II — Heavenward Howl — When you next cast a creature spell this turn, that creature enters with an additional +1/+1 counter on it.",
					Chapters: []int{2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateDelayedTrigger{
											Trigger: game.DelayedTriggerDef{
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.CreateReplacement{
																		Replacement: &game.ReplacementEffect{
																			Description: "That creature enters with an additional +1/+1 counter on it.",
																			Duration:    game.DurationPermanent,
																			MatchEvent:  game.EventPermanentEnteredBattlefield,
																			EntersWithCounters: []game.CounterPlacement{
																				game.CounterPlacement{
																					Kind:   counter.PlusOnePlusOne,
																					Amount: 1,
																				},
																			},
																		},
																		Duration: game.DurationUntilEndOfTurn,
																		Object:   game.EventStackObjectReference(),
																	},
																},
															},
														},
													},
													MinModes: 1,
													MaxModes: 1,
												},
												EventPattern: opt.Val(game.TriggerPattern{
													Event:           game.EventSpellCast,
													Controller:      game.TriggerControllerYou,
													CardSelection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
													StackObjectKind: game.StackSpell,
													CounterKind:     counter.PlusOnePlusOne,
												}),
												OneShot: true,
												Window:  game.DelayedWindowThisTurn,
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
				game.ChapterAbility{
					Text:     "III — Ecliptic Growl — Draw a card if you control the creature with the greatest power or tied for the greatest power.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                                    "if you control the creature with the greatest power or tied for the greatest power",
												SourceCounterKind:                       counter.PlusOnePlusOne,
												ControllerControlsGreatestPowerCreature: true,
											}),
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
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
			I — Crescent Fang — Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
			II — Heavenward Howl — When you next cast a creature spell this turn, that creature enters with an additional +1/+1 counter on it.
			III — Ecliptic Growl — Draw a card if you control the creature with the greatest power or tied for the greatest power.
		`,
		},
	}
}
