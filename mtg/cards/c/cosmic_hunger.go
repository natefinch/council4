package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Cosmic Hunger
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Target creature you control deals damage equal to its power to another target creature, planeswalker, or battle.
var CosmicHunger = &game.CardDef{
	Name: "Cosmic Hunger",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     2,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []game.CardType{game.TypeInstant},
	OracleText:    "Target creature you control deals damage equal to its power to another target creature, planeswalker, or battle.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Target creature you control deals damage equal to its power to another target creature, planeswalker, or battle.",
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []game.CardType{game.TypeCreature},
						Controller:     game.ControllerYou,
					},
				},
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "another creature, planeswalker, or battle",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []game.CardType{game.TypeCreature, game.TypePlaneswalker, game.TypeBattle},
						Another:        true,
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
	},
}
