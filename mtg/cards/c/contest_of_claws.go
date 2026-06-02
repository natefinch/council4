package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ContestOfClaws is the card definition for Contest of Claws.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Target creature you control deals damage equal to its power to another target creature. If excess damage was dealt this way, discover X, where X is that excess damage.
var ContestOfClaws = &game.CardDef{
	Name: "Contest of Claws",
	ManaCost: opt.Val(cost.Mana{
		cost.O(1),
		cost.G,
	}),
	Colors:        []color.Color{color.Green},
	ColorIdentity: mana.NewColorIdentity(color.Green),
	Types:         []types.Card{types.Sorcery},
	OracleText:    "Target creature you control deals damage equal to its power to another target creature. If excess damage was dealt this way, discover X, where X is that excess damage. (Exile cards from the top of your library until you exile a nonland card with that mana value or less. Cast it without paying its mana cost or put it into your hand. Put the rest on the bottom in a random order.)",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Target creature you control deals damage equal to its power to another target creature. If excess damage was dealt this way, discover X, where X is that excess damage.",
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
					Constraint: "another target creature",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
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
					ResultAmount: game.EffectResultAmountExcessDamage,
					LinkID:       "excess",
				},
				{
					Type:        game.EffectDiscover,
					TargetIndex: game.TargetIndexController,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:   game.DynamicAmountPreviousEffectExcessDamage,
						LinkID: "excess",
					}),
					ResultCondition: opt.Val(game.EffectResultCondition{
						LinkID:    "excess",
						Succeeded: game.TriTrue,
					}),
				},
			},
		},
	},
}
