package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Chandra's Ignition
//
// Type: Sorcery
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Target creature you control deals damage equal to its power to each other creature and each opponent.
//
// Missing primitives:
//   - No EffectSelectorAllCreaturesExceptTarget / "each other creature" selector;
//     EffectSelectorAllCreatures is used but incorrectly includes the targeting
//     creature itself. ImplementationID "chandras-ignition" must exclude it.
//   - No EffectSelectorAllOpponents selector for the "each opponent" clause.
//     ImplementationID "chandras-ignition" must add a separate pass damaging
//     each opponent for the same dynamic amount.
var ChandraSIgnition = &game.CardDef{
	Name: "Chandra's Ignition",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(3),
		mana.ColoredMana(mana.Red),
		mana.ColoredMana(mana.Red),
	}),
	ManaValue:        5,
	Colors:           []mana.Color{mana.Red},
	ColorIdentity:    mana.NewColorIdentity(mana.Red),
	Types:            []game.CardType{game.TypeSorcery},
	OracleText:       "Target creature you control deals damage equal to its power to each other creature and each opponent.",
	ImplementationID: "chandras-ignition",
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
						PermanentTypes: []game.CardType{game.TypeCreature},
						Controller:     game.ControllerYou,
					},
				},
			},
			Effects: []game.Effect{
				{
					// Approximation: should be "each OTHER creature" (excluding target 0)
					// and also hit each opponent. ImplementationID handles both corrections.
					Type:     game.EffectDamage,
					Selector: game.EffectSelectorAllCreatures,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:        game.DynamicAmountTargetPower,
						TargetIndex: 0,
					}),
					Description: "deals damage equal to its power to each other creature and each opponent",
				},
			},
		},
	},
}
