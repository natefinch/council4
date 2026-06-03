package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MasterSRebuke is the card definition for Master's Rebuke.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Target creature you control deals damage equal to its power to target creature or planeswalker you don't control.
var MasterSRebuke = &game.CardDef{CardFace: game.CardFace{Name: "Master's Rebuke",
	ManaCost: opt.Val(cost.Mana{
		cost.O(1),
		cost.G,
	}),
	Colors: []color.Color{color.Green},

	Types:      []types.Card{types.Instant},
	OracleText: "Target creature you control deals damage equal to its power to target creature or planeswalker you don't control.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Target creature you control deals damage equal to its power to target creature or planeswalker you don't control.",
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
					Constraint: "creature or planeswalker you don't control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature, types.Planeswalker},
						Controller:     game.ControllerOpponent,
					},
				},
			},
			Effects: []game.Effect{
				{
					Type: game.EffectDamage,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:        game.DynamicAmountTargetPower,
						TargetIndex: 0,
					}),
					TargetIndex: 1,
				},
			},
		},
	}}, ColorIdentity: color.NewIdentity(color.Green),
}
