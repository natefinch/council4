package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Fiendlash is the card definition for Fiendlash.
//
// Type: Artifact — Equipment
// Cost: {1}{R}
//
// Oracle text:
//
//	Equipped creature gets +2/+0 and has reach.
//	Whenever equipped creature is dealt damage, it deals damage equal to its power to target player or planeswalker.
//	Equip {2}{R}
var Fiendlash = &game.CardDef{
	Name: "Fiendlash",
	ManaCost: opt.Val(cost.Mana{
		cost.O(1),
		cost.R,
	}),
	Colors:        []color.Color{color.Red},
	ColorIdentity: mana.NewColorIdentity(color.Red),
	Types:         []types.Card{types.Artifact},
	Subtypes:      []types.Sub{types.Equipment},
	OracleText:    "Equipped creature gets +2/+0 and has reach.\nWhenever equipped creature is dealt damage, it deals damage equal to its power to target player or planeswalker.\nEquip {2}{R}",
	Abilities: []game.AbilityDef{
		{
			Kind: game.StaticAbility,
			Text: "Equipped creature gets +2/+0 and has reach.",
			Effects: []game.Effect{
				{
					Type:        game.EffectApplyContinuous,
					TargetIndex: game.TargetIndexSourcePermanent,
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:          game.LayerPowerToughnessModify,
							Selector:       game.EffectSelectorEquippedCreature,
							PowerDelta:     2,
							ToughnessDelta: 0,
						},
						{
							Layer:       game.LayerAbility,
							Selector:    game.EffectSelectorEquippedCreature,
							AddKeywords: []game.Keyword{game.Reach},
						},
					},
				},
			},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever equipped creature is dealt damage, it deals damage equal to its power to target player or planeswalker.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:           game.EventDamageDealt,
					Source:          game.TriggerSourceAttachedPermanent,
					DamageRecipient: game.DamageRecipientPermanent,
				},
			}),
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "player or planeswalker",
					Allow:      game.TargetAllowPlayer | game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Planeswalker},
					},
				},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectDamage,
					TargetIndex: 0,
					DamageSource: opt.Val(game.ObjectReference{
						Kind:        game.ObjectReferenceAttachedPermanent,
						TargetIndex: game.TargetIndexSourcePermanent,
					}),
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind: game.DynamicAmountObjectPower,
						Object: game.ObjectReference{
							Kind:        game.ObjectReferenceAttachedPermanent,
							TargetIndex: game.TargetIndexSourcePermanent,
						},
					}),
				},
			},
		},
		{
			Kind:     game.ActivatedAbility,
			Text:     "Equip {2}{R}",
			Keywords: []game.Keyword{game.Equip},
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Timing: game.SorceryOnly,
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
			},
		},
	},
}
