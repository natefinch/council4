package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KardurSViciousReturn is the card definition for Kardur's Vicious Return.
//
// Type: Enchantment — Saga
// Cost: {2}{B}{R}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I — You may sacrifice a creature. When you do, this Saga deals 3 damage to any target.
//	II — Each player discards a card.
//	III — Return target creature card from your graveyard to the battlefield. Put a +1/+1 counter on it. It gains haste until your next turn.
var KardurSViciousReturn = newKardurSViciousReturn

func newKardurSViciousReturn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Kardur's Vicious Return",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.R,
			}),
			Colors:   []color.Color{color.Black, color.Red},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — You may sacrifice a creature. When you do, this Saga deals 3 damage to any target.",
					Chapters: []int{1},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.ControllerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.CreateReflexiveTrigger{
									Trigger: game.ReflexiveTriggerDef{
										Content: game.Mode{
											Targets: []game.TargetSpec{
												game.TargetSpec{
													MinTargets: 1,
													MaxTargets: 1,
													Constraint: "any target",
													Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
												},
											},
											Sequence: []game.Instruction{
												{
													Primitive: game.Damage{
														Amount:       game.Fixed(3),
														Recipient:    game.AnyTargetDamageRecipient(0),
														DamageSource: opt.Val(game.SourcePermanentReference()),
													},
												},
											},
										}.Ability(),
									},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
				game.ChapterAbility{
					Text:     "II — Each player discards a card.",
					Chapters: []int{2},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount:      game.Fixed(1),
									PlayerGroup: game.AllPlayersReference(),
								},
							},
						},
					}.Ability(),
				},
				game.ChapterAbility{
					Text:     "III — Return target creature card from your graveyard to the battlefield. Put a +1/+1 counter on it. It gains haste until your next turn.",
					Chapters: []int{3},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
									PublishLinked: game.LinkedKey("leave-bf-exile-1"),
								},
							},
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.LinkedObjectReference("leave-bf-exile-1"),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Haste,
											},
										},
									},
									Duration: game.DurationUntilYourNextTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
			I — You may sacrifice a creature. When you do, this Saga deals 3 damage to any target.
			II — Each player discards a card.
			III — Return target creature card from your graveyard to the battlefield. Put a +1/+1 counter on it. It gains haste until your next turn.
		`,
		},
	}
}
