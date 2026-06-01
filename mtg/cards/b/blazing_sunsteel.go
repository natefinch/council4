package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Blazing Sunsteel
//
// Type: Artifact — Equipment
// Cost: {1}{R}
//
// Oracle text:
//
//	Equipped creature gets +1/+0 for each opponent you have.
//	Whenever equipped creature is dealt damage, it deals that much damage to any target.
//	Equip {4}
var BlazingSunsteel = &game.CardDef{
	Name: "Blazing Sunsteel",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Red),
	}),
	ManaValue:     2,
	Colors:        []mana.Color{mana.Red},
	ColorIdentity: mana.NewColorIdentity(mana.Red),
	Types:         []types.Card{types.Artifact},
	Subtypes:      []types.Sub{types.Equipment},
	OracleText:    "Equipped creature gets +1/+0 for each opponent you have.\nWhenever equipped creature is dealt damage, it deals that much damage to any target.\nEquip {4}",
	Abilities: []game.AbilityDef{
		{
			Kind: game.StaticAbility,
			Text: "Equipped creature gets +1/+0 for each opponent you have.",
			Effects: []game.Effect{
				{
					Type:        game.EffectModifyPT,
					TargetIndex: -2,
					Selector:    game.EffectSelectorEquippedCreature,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind: game.DynamicAmountOpponentCount,
					}),
				},
			},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever equipped creature is dealt damage, it deals that much damage to any target.",
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
					Constraint: "any target",
					Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
				},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectDamage,
					TargetIndex: 0,
					DamageSource: opt.Val(game.ObjectReference{
						Kind:        game.ObjectReferenceAttachedPermanent,
						TargetIndex: -1,
					}),
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind: game.DynamicAmountEventDamage,
					}),
				},
			},
		},
		{
			Kind:     game.ActivatedAbility,
			Text:     "Equip {4}",
			Keywords: []game.Keyword{game.Equip},
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(4),
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
