package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
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
var Arena = &game.CardDef{
	Name:       "Arena",
	ManaValue:  0,
	Types:      []types.Card{types.Land},
	OracleText: "{3}, {T}: Tap target creature you control and target creature of an opponent's choice they control. Those creatures fight each other. (Each deals damage equal to its power to the other.)",
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
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerYou,
					},
				},
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature of an opponent's choice they control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerYou,
					},
					Chooser: game.TargetChooserOpponent,
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
