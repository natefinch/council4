package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Arena
//
// Type: Land
//
// Oracle text:
//
//	{3}, {T}: Tap target creature you control and target creature of an opponent's
//	choice they control. Those creatures fight each other.
//
// Missing primitives:
//   - TargetSpec has no "chooser" field; there is no way to declare that the
//     second target is chosen by an opponent rather than the active player.
//     ImplementationID "arena" is set so a hand-written rules handler can prompt
//     the correct player to choose the opponent-controlled creature.
var Arena = &game.CardDef{
	Name:             "Arena",
	ManaValue:        0,
	Types:            []game.CardType{game.TypeLand},
	OracleText:       "{3}, {T}: Tap target creature you control and target creature of an opponent's choice they control. Those creatures fight each other. (Each deals damage equal to its power to the other.)",
	ImplementationID: "arena",
	Abilities: []game.AbilityDef{
		{
			Kind: game.ActivatedAbility,
			Text: "{3}, {T}: Tap target creature you control and target creature of an opponent's choice they control. Those creatures fight each other.",
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(3),
			}),
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
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
					// Opponent chooses which of their creatures to target.
					// The "opponent chooses this target" mechanic is not yet
					// representable; ImplementationID "arena" gates on this.
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature of an opponent's choice they control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []game.CardType{game.TypeCreature},
						Controller:     game.ControllerOpponent,
					},
				},
			},
			Effects: []game.Effect{
				{Type: game.EffectTap, TargetIndex: 0},
				{Type: game.EffectTap, TargetIndex: 1},
				{Type: game.EffectFight},
			},
		},
	},
}
