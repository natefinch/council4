package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HunterSTalent is the card definition for Hunter's Talent.
//
// Type: Enchantment — Class
// Cost: {1}{G}
//
// Oracle text:
//
//	(Gain the next level as a sorcery to add its ability.)
//	When this Class enters, target creature you control deals damage equal to its power to target creature you don't control.
//	{1}{G}: Level 2
//	Whenever you attack, target attacking creature gets +1/+0 and gains trample until end of turn.
//	{3}{G}: Level 3
//	At the beginning of your end step, if you control a creature with power 4 or greater, draw a card.
var HunterSTalent = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hunter's Talent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Class},
			OracleText: `
				(Gain the next level as a sorcery to add its ability.)
				When this Class enters, target creature you control deals damage equal to its power to target creature you don't control.
				{1}{G}: Level 2
				Whenever you attack, target attacking creature gets +1/+0 and gains trample until end of turn.
				{3}{G}: Level 3
				At the beginning of your end step, if you control a creature with power 4 or greater, draw a card.
			`,
		},
	}

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbility{
			Text: `
				When this Class enters, target creature you control deals damage equal to its power to target creature you don't control.
			`,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:  game.EventPermanentEnteredBattlefield,
					Source: game.TriggerSourceSelf,
				},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature you control",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{
								types.Creature,
							},
							Controller: game.ControllerYou,
						},
					},
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature you don't control",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{
								types.Creature,
							},
							Controller: game.ControllerOpponent,
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:        game.DynamicAmountTargetPower,
								TargetIndex: 0,
							}),
							Recipient: game.TargetRecipient(1),
							DamageSource: opt.Val(game.ObjectReference{
								Kind:        game.ObjectReferenceTargetPermanent,
								TargetIndex: 0,
							}),
						},
					},
				},
			}.Ability(),
		},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.ActivatedAbility{
			Text: `
				{1}{G}: Level 2
			`,
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Timing: game.SorceryOnly,
			ActivationCondition: opt.Val(game.Condition{
				SourceClassLevelLessThan: 2,
			}),
			Content: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SetClassLevel{
							TargetIndex: game.TargetIndexSourcePermanent,
							Amount:      game.Fixed(2),
						},
					},
				},
			}.Ability(),
		},
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbility{
			Text: `
				Whenever you attack, target attacking creature gets +1/+0 and gains trample until end of turn.
			`,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:      game.EventAttackerDeclared,
					Controller: game.TriggerControllerYou,
				},
				InterveningCondition: opt.Val(game.Condition{
					SourceClassLevelAtLeast: 2,
				}),
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "attacking creature",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{
								types.Creature,
							},
							Controller:  game.ControllerYou,
							CombatState: game.CombatStateAttacking,
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							TargetIndex: 0,
							PowerDelta:  game.Fixed(1),
							Duration:    game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							TargetIndex: 0,
							ContinuousEffects: []game.ContinuousEffect{
								{
									Layer: game.LayerAbility,
									AddKeywords: []game.Keyword{
										game.Trample,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability(),
		},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.ActivatedAbility{
			Text: `
				{3}{G}: Level 3
			`,
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Timing: game.SorceryOnly,
			ActivationCondition: opt.Val(game.Condition{
				SourceClassLevelAtLeast:  2,
				SourceClassLevelLessThan: 3,
			}),
			Content: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SetClassLevel{
							TargetIndex: game.TargetIndexSourcePermanent,
							Amount:      game.Fixed(3),
						},
					},
				},
			}.Ability(),
		},
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbility{
			Text: `
				At the beginning of your end step, if you control a creature with power 4 or greater, draw a card.
			`,
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:      game.EventBeginningOfStep,
					Controller: game.TriggerControllerYou,
					Step:       game.StepEnd,
				},
				InterveningIf: "if you control a creature with power 4 or greater",
				InterveningCondition: opt.Val(game.Condition{
					Text: "if you control a creature with power 4 or greater",
					ControllerControls: game.PermanentFilter{
						Types: []types.Card{
							types.Creature,
						},
						Power: opt.Val(compare.Int{
							Op:    compare.GreaterOrEqual,
							Value: 4,
						}),
					},
					SourceClassLevelAtLeast: 3,
				}),
			},
			Content: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount:      game.Fixed(1),
							TargetIndex: game.TargetIndexController,
						},
					},
				},
			}.Ability(),
		},
	)
	return card
}()
