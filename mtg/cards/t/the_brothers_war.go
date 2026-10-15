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

// TheBrothersWar is the card definition for The Brothers' War.
//
// Type: Enchantment — Saga
// Cost: {3}{R}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I — Create two tapped Powerstone tokens.
//	II — Choose two target players. Until your next turn, each creature they control attacks the other chosen player each combat if able.
//	III — This Saga deals X damage to any target and X damage to any other target, where X is the number of artifacts you control.
var TheBrothersWar = newTheBrothersWar

func newTheBrothersWar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "The Brothers' War",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — Create two tapped Powerstone tokens.",
					Chapters: []int{1},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:      game.Fixed(2),
											Source:      game.TokenDef(theBrothersWarToken),
											EntryTapped: true,
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
					Text:     "II — Choose two target players. Until your next turn, each creature they control attacks the other chosen player each combat if able.",
					Chapters: []int{2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "player",
										Allow:      game.TargetAllowPlayer,
									},
									game.TargetSpec{
										MinTargets:               1,
										MaxTargets:               1,
										Constraint:               "player",
										Allow:                    game.TargetAllowPlayer,
										DistinctFromPriorTargets: true,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyRule{
											RuleEffects: []game.RuleEffect{
												game.RuleEffect{
													Kind:                    game.RuleEffectMustAttack,
													Duration:                game.DurationPermanent,
													PermanentTypes:          []types.Card{types.Creature},
													AffectedPlayerRef:       game.TargetPlayerReference(0),
													RequiredAttackTargetRef: game.TargetPlayerReference(1),
												},
												game.RuleEffect{
													Kind:                    game.RuleEffectMustAttack,
													Duration:                game.DurationPermanent,
													PermanentTypes:          []types.Card{types.Creature},
													AffectedPlayerRef:       game.TargetPlayerReference(1),
													RequiredAttackTargetRef: game.TargetPlayerReference(0),
												},
											},
											Duration: game.DurationUntilYourNextTurn,
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
					Text:     "III — This Saga deals X damage to any target and X damage to any other target, where X is the number of artifacts you control.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "any target",
										Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
									},
									game.TargetSpec{
										MinTargets:               1,
										MaxTargets:               1,
										Constraint:               "any other target",
										Allow:                    game.TargetAllowPermanent | game.TargetAllowPlayer,
										Selection:                opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
										DistinctFromPriorTargets: true,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou}),
											}),
											Recipient:    game.AnyTargetDamageRecipient(0),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou}),
											}),
											Recipient:    game.AnyTargetDamageRecipient(1),
											DamageSource: opt.Val(game.SourcePermanentReference()),
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
			I — Create two tapped Powerstone tokens.
			II — Choose two target players. Until your next turn, each creature they control attacks the other chosen player each combat if able.
			III — This Saga deals X damage to any target and X damage to any other target, where X is the number of artifacts you control.
		`,
		},
	}
}

var theBrothersWarToken = newTheBrothersWarToken()

func newTheBrothersWarToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Powerstone",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Powerstone},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.C,
											SpendRider: opt.Val(game.ManaSpendRider{
												Condition:   game.ManaSpendCastArtifactSpell,
												Restriction: game.ManaSpendRestrictedToCondition,
											}),
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
