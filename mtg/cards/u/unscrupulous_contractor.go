package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnscrupulousContractor is the card definition for Unscrupulous Contractor.
//
// Type: Creature — Human Assassin
// Cost: {2}{B}
//
// Oracle text:
//
//	When this creature enters, you may sacrifice a creature. When you do, target player draws two cards and loses 2 life.
//	Plot {2}{B} (You may pay {2}{B} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
var UnscrupulousContractor = newUnscrupulousContractor

func newUnscrupulousContractor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Unscrupulous Contractor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Assassin},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Plot {2}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.PlotKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.B,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, you may sacrifice a creature. When you do, target player draws two cards and loses 2 life.",
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
										Primitive: game.SacrificePermanents{
											Player:    game.ControllerReference(),
											Amount:    game.Fixed(1),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
										},
										Optional:      true,
										PublishResult: game.ResultKey("if-you-do"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("if-you-do")},
										},
									},
									game.Instruction{
										Primitive: game.CreateReflexiveTrigger{
											Trigger: game.ReflexiveTriggerDef{
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Targets: []game.TargetSpec{
																game.TargetSpec{
																	MinTargets: 1,
																	MaxTargets: 1,
																	Constraint: "target player",
																	Allow:      game.TargetAllowPlayer,
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.Draw{
																		Amount: game.Fixed(2),
																		Player: game.TargetPlayerReference(0),
																	},
																},
																game.Instruction{
																	Primitive: game.LoseLife{
																		Amount: game.Fixed(2),
																		Player: game.TargetPlayerReference(0),
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
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("if-you-do"),
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
			When this creature enters, you may sacrifice a creature. When you do, target player draws two cards and loses 2 life.
			Plot {2}{B} (You may pay {2}{B} and exile this card from your hand. Cast it as a sorcery on a later turn without paying its mana cost. Plot only as a sorcery.)
		`,
		},
	}
}
