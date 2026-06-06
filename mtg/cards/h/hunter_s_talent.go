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
		game.TriggeredAbilityBody{
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
			Content: game.PlainAbilityContent{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature you control",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Creature},
							Controller:     game.ControllerYou,
						},
					},
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature you don't control",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Creature},
							Controller:     game.ControllerOpponent,
						},
					},
				},
				Sequence: []game.Effect{
					{
						Type:        game.EffectDamage,
						TargetIndex: 1,
						DamageSource: opt.Val(game.ObjectReference{
							Kind:        game.ObjectReferenceTargetPermanent,
							TargetIndex: 0,
						}),
						DynamicAmount: opt.Val(game.DynamicAmount{
							Kind:        game.DynamicAmountTargetPower,
							TargetIndex: 0,
						}),
					},
				},
			},
		},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.ActivatedAbilityBody{
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
			Content: game.PlainAbilityContent{
				Sequence: []game.Effect{
					{Type: game.EffectSetClassLevel, Amount: 2, TargetIndex: game.TargetIndexSourcePermanent},
				},
			},
		},
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbilityBody{
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
			Content: game.PlainAbilityContent{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "attacking creature",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Creature},
							Controller:     game.ControllerYou,
							CombatState:    game.CombatStateAttacking,
						},
					},
				},
				Sequence: []game.Effect{
					{
						Type:           game.EffectModifyPT,
						TargetIndex:    0,
						PowerDelta:     1,
						ToughnessDelta: 0,
						UntilEndOfTurn: true,
					},
					{
						Type:           game.EffectApplyContinuous,
						TargetIndex:    0,
						UntilEndOfTurn: true,
						ContinuousEffects: []game.ContinuousEffect{
							{
								Layer:       game.LayerAbility,
								AddKeywords: []game.Keyword{game.Trample},
							},
						},
					},
				},
			},
		},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.ActivatedAbilityBody{
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
			Content: game.PlainAbilityContent{
				Sequence: []game.Effect{
					{Type: game.EffectSetClassLevel, Amount: 3, TargetIndex: game.TargetIndexSourcePermanent},
				},
			},
		},
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbilityBody{
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
						Types: []types.Card{types.Creature},
						Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
					},
					SourceClassLevelAtLeast: 3,
				}),
			},
			Content: game.PlainAbilityContent{
				Sequence: []game.Effect{
					{Type: game.EffectDraw, Amount: 1, TargetIndex: game.TargetIndexController},
				},
			},
		},
	)
	return card
}()
