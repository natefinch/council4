package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SummonPrimalOdin is the card definition for Summon: Primal Odin.
//
// Type: Enchantment Creature — Saga Knight
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I — Gungnir — Destroy target creature an opponent controls.
//	II — Zantetsuken — This creature gains "Whenever this creature deals combat damage to a player, that player loses the game."
//	III — Hall of Sorrow — Draw two cards. Each player loses 2 life.
var SummonPrimalOdin = newSummonPrimalOdin

func newSummonPrimalOdin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Summon: Primal Odin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Saga, types.Knight},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — Gungnir — Destroy target creature an opponent controls.",
					Chapters: []int{1},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature an opponent controls",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
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
					Text:     "II — Zantetsuken — This creature gains \"Whenever this creature deals combat damage to a player, that player loses the game.\"",
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
															Text: "Whenever this creature deals combat damage to a player, that player loses the game.",
															Trigger: game.TriggerCondition{
																Type: game.TriggerWhenever,
																Pattern: game.TriggerPattern{
																	Event:               game.EventDamageDealt,
																	Source:              game.TriggerSourceSelf,
																	Subject:             game.TriggerSubjectDamageSource,
																	StackObjectKind:     game.StackSpell,
																	DamageRecipient:     game.DamageRecipientPlayer,
																	RequireCombatDamage: true,
																	CounterKind:         counter.PlusOnePlusOne,
																},
															},
															Content: game.AbilityContent{
																Modes: []game.Mode{
																	game.Mode{
																		Sequence: []game.Instruction{
																			game.Instruction{
																				Primitive: game.PlayerLosesGame{
																					Player: game.EventPlayerReference(),
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
					Text:     "III — Hall of Sorrow — Draw two cards. Each player loses 2 life.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(2),
											PlayerGroup: game.AllPlayersReference(),
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
			I — Gungnir — Destroy target creature an opponent controls.
			II — Zantetsuken — This creature gains "Whenever this creature deals combat damage to a player, that player loses the game."
			III — Hall of Sorrow — Draw two cards. Each player loses 2 life.
		`,
		},
	}
}
