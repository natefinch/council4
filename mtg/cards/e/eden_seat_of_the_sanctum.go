package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EdenSeatOfTheSanctum is the card definition for Eden, Seat of the Sanctum.
//
// Type: Land — Town
//
// Oracle text:
//
//	{T}: Add {C}.
//	{5}, {T}: Mill two cards. Then you may sacrifice this land. When you do, return another target permanent card from your graveyard to your hand.
var EdenSeatOfTheSanctum = newEdenSeatOfTheSanctum

func newEdenSeatOfTheSanctum() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Eden, Seat of the Sanctum",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Town},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}, {T}: Mill two cards. Then you may sacrifice this land. When you do, return another target permanent card from your graveyard to your hand.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:      game.Fixed(2),
											Player:      game.ControllerReference(),
											Destination: zone.Graveyard,
										},
									},
									game.Instruction{
										Primitive: game.Sacrifice{
											Object: game.SourcePermanentReference(),
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
																	Constraint: "another target permanent card from your graveyard",
																	Allow:      game.TargetAllowCard,
																	Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, Controller: game.ControllerYou}),
																	TargetZone: zone.Graveyard,
																},
															},
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.MoveCard{
																		Card: game.CardReference{
																			Kind: game.CardReferenceTarget,
																		},
																		FromZone:    zone.Graveyard,
																		Destination: zone.Hand,
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{5}, {T}: Mill two cards. Then you may sacrifice this land. When you do, return another target permanent card from your graveyard to your hand.
		`,
		},
	}
}
