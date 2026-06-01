package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Bite Down
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Target creature you control deals damage equal to its power to target creature
//	or planeswalker you don't control.
//
// Note: EffectDamage attributes damage to the spell source (r.obj), not to
// the dealing creature. Lifelink and deathtouch on target 0 will therefore
// not trigger via this effect. A future "creature-sourced damage" primitive
// (or ImplementationID) would be needed for full rules accuracy.

var BiteDown = &game.CardDef{
	Name: "Bite Down",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     2,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []game.CardType{game.TypeInstant},
	OracleText:    "Target creature you control deals damage equal to its power to target creature or planeswalker you don't control.",
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
						PermanentTypes: []game.CardType{game.TypeCreature},
						Controller:     game.ControllerYou,
					},
				},
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature or planeswalker you don't control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []game.CardType{game.TypeCreature, game.TypePlaneswalker},
						Controller:     game.ControllerOpponent,
					},
				},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectDamage,
					TargetIndex: 1,
					DynamicAmount: opt.Val(game.DynamicAmount{
						Kind:        game.DynamicAmountTargetPower,
						TargetIndex: 0,
					}),
				},
			},
		},
	},
}
