package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChandraSIgnition is the card definition for Chandra's Ignition.
//
// Type: Sorcery
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Target creature you control deals damage equal to its power to each other creature and each opponent.
var ChandraSIgnition = &game.CardDef{CardFace: game.CardFace{Name: "Chandra's Ignition",
	ManaCost: opt.Val(cost.Mana{
		cost.O(3),
		cost.R,
		cost.R,
	}),
	Colors: []color.Color{color.Red},

	Types:      []types.Card{types.Sorcery},
	OracleText: "Target creature you control deals damage equal to its power to each other creature and each opponent.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Target creature you control deals damage equal to its power to each other creature and each opponent.",
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
			Effects: []game.Effect{
				{
					Type:        game.EffectDamage,
					TargetIndex: 0,
					Selector:    game.EffectSelectorAllCreaturesExceptTarget,
					DamageSource: opt.Val(game.ObjectReference{
						Kind:        game.ObjectReferenceTargetPermanent,
						TargetIndex: 0,
					}),
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:        game.DynamicAmountTargetPower,
						TargetIndex: 0,
					}),
					Description: "deals damage equal to its power to each other creature",
				},
				{
					Type:           game.EffectDamage,
					PlayerSelector: game.PlayerSelectorOpponents,
					DamageSource: opt.Val(game.ObjectReference{
						Kind:        game.ObjectReferenceTargetPermanent,
						TargetIndex: 0,
					}),
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:        game.DynamicAmountTargetPower,
						TargetIndex: 0,
					}),
					Description: "deals damage equal to its power to each opponent",
				},
			},
		},
	}}, ColorIdentity: color.NewIdentity(color.Red),
}
