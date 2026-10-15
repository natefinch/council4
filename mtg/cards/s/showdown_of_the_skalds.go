package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShowdownOfTheSkalds is the card definition for Showdown of the Skalds.
//
// Type: Enchantment — Saga
// Cost: {2}{R}{W}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I — Exile the top four cards of your library. Until the end of your next turn, you may play those cards.
//	II, III — Whenever you cast a spell this turn, put a +1/+1 counter on target creature you control.
var ShowdownOfTheSkalds = newShowdownOfTheSkalds

func newShowdownOfTheSkalds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Showdown of the Skalds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.W,
			}),
			Colors:   []color.Color{color.Red, color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — Exile the top four cards of your library. Until the end of your next turn, you may play those cards.",
					Chapters: []int{1},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ImpulseExile{
											Player:   game.ControllerReference(),
											Amount:   game.Fixed(4),
											Duration: game.DurationUntilEndOfYourNextTurn,
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
					Text:     "II, III — Whenever you cast a spell this turn, put a +1/+1 counter on target creature you control.",
					Chapters: []int{2, 3},
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
															Targets: []game.TargetSpec{
																game.TargetSpec{
																	MinTargets: 1,
																	MaxTargets: 1,
																	Constraint: "target creature you control",
																	Allow:      game.TargetAllowPermanent,
																	Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
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
															},
														},
													},
													MinModes: 1,
													MaxModes: 1,
												},
												EventPattern: opt.Val(game.TriggerPattern{
													Event:           game.EventSpellCast,
													Controller:      game.TriggerControllerYou,
													StackObjectKind: game.StackSpell,
													CounterKind:     counter.PlusOnePlusOne,
												}),
												Window: game.DelayedWindowThisTurn,
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
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
			I — Exile the top four cards of your library. Until the end of your next turn, you may play those cards.
			II, III — Whenever you cast a spell this turn, put a +1/+1 counter on target creature you control.
		`,
		},
	}
}
