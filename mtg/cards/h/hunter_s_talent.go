package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
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
var HunterSTalent = &game.CardDef{
	Name: "Hunter's Talent",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     2,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []types.Card{types.Enchantment},
	Subtypes:      []types.Sub{types.Class},
	OracleText:    "(Gain the next level as a sorcery to add its ability.)\nWhen this Class enters, target creature you control deals damage equal to its power to target creature you don't control.\n{1}{G}: Level 2\nWhenever you attack, target attacking creature gets +1/+0 and gains trample until end of turn.\n{3}{G}: Level 3\nAt the beginning of your end step, if you control a creature with power 4 or greater, draw a card.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.TriggeredAbility,
			Text: "When this Class enters, target creature you control deals damage equal to its power to target creature you don't control.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:  game.EventPermanentEnteredBattlefield,
					Source: game.TriggerSourceSelf,
				},
			}),
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
			Effects: []game.Effect{
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
		{
			Kind: game.ActivatedAbility,
			Text: "{1}{G}: Level 2",
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(1),
				mana.ColoredMana(mana.Green),
			}),
			Timing: game.SorceryOnly,
			ActivationCondition: opt.Val(game.Condition{
				SourceClassLevelLessThan: 2,
			}),
			Effects: []game.Effect{
				{Type: game.EffectSetClassLevel, Amount: 2, TargetIndex: game.TargetIndexSourcePermanent},
			},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever you attack, target attacking creature gets +1/+0 and gains trample until end of turn.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:      game.EventAttackerDeclared,
					Controller: game.TriggerControllerYou,
				},
				InterveningCondition: opt.Val(game.Condition{
					SourceClassLevelAtLeast: 2,
				}),
			}),
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
			Effects: []game.Effect{
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
		{
			Kind: game.ActivatedAbility,
			Text: "{3}{G}: Level 3",
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(3),
				mana.ColoredMana(mana.Green),
			}),
			Timing: game.SorceryOnly,
			ActivationCondition: opt.Val(game.Condition{
				SourceClassLevelAtLeast:  2,
				SourceClassLevelLessThan: 3,
			}),
			Effects: []game.Effect{
				{Type: game.EffectSetClassLevel, Amount: 3, TargetIndex: game.TargetIndexSourcePermanent},
			},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "At the beginning of your end step, if you control a creature with power 4 or greater, draw a card.",
			Trigger: opt.Val(game.TriggerCondition{
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
			}),
			Effects: []game.Effect{
				{Type: game.EffectDraw, Amount: 1, TargetIndex: game.TargetIndexController},
			},
		},
	},
}
