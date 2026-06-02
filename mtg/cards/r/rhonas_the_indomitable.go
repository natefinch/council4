package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RhonasTheIndomitable is the card definition for Rhonas the Indomitable.
//
// Type: Legendary Creature — God
// Cost: {2}{G}
//
// Oracle text:
//
//	Deathtouch, indestructible
//	Rhonas can't attack or block unless you control another creature with power 4 or greater.
//	{2}{G}: Another target creature gets +2/+0 and gains trample until end of turn.
var RhonasTheIndomitable = &game.CardDef{
	Name: "Rhonas the Indomitable",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(2),
		mana.ColoredMana(mana.Green),
	}),
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Supertypes:    []types.Super{types.Legendary},
	Types:         []types.Card{types.Creature},
	Subtypes:      []types.Sub{types.Sub("God")},
	Power:         opt.Val(game.PT{Value: 5}),
	Toughness:     opt.Val(game.PT{Value: 5}),
	OracleText:    "Deathtouch, indestructible\nRhonas can't attack or block unless you control another creature with power 4 or greater.\n{2}{G}: Another target creature gets +2/+0 and gains trample until end of turn.",
	Abilities: []game.AbilityDef{
		{
			Kind:     game.StaticAbility,
			Text:     "Deathtouch, indestructible",
			Keywords: []game.Keyword{game.Deathtouch, game.Indestructible},
		},
		{
			Kind: game.StaticAbility,
			Text: "Rhonas can't attack or block unless you control another creature with power 4 or greater.",
			Condition: opt.Val(game.Condition{
				Text:   "unless you control another creature with power 4 or greater",
				Negate: true,
				ControllerControls: game.PermanentFilter{
					Types:         []types.Card{types.Creature},
					Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
					ExcludeSource: true,
				},
			}),
			Effects: []game.Effect{
				{
					Type:        game.EffectApplyRule,
					TargetIndex: game.TargetIndexController,
					RuleEffects: []game.RuleEffect{
						{Kind: game.RuleEffectCantAttack, AffectedSource: true},
						{Kind: game.RuleEffectCantBlock, AffectedSource: true},
					},
				},
			},
		},
		{
			Kind: game.ActivatedAbility,
			Text: "{2}{G}: Another target creature gets +2/+0 and gains trample until end of turn.",
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(2),
				mana.ColoredMana(mana.Green),
			}),
			Targets: []game.TargetSpec{
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
					Type:           game.EffectModifyPT,
					TargetIndex:    0,
					PowerDelta:     2,
					ToughnessDelta: 0,
					UntilEndOfTurn: true,
				},
				{
					Type:           game.EffectApplyContinuous,
					TargetIndex:    0,
					UntilEndOfTurn: true,
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Trample},
						},
					},
				},
			},
		},
	},
}
