package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HazoretGodseeker is the card definition for Hazoret, Godseeker.
//
// Type: Legendary Creature — God
// Cost: {1}{R}
//
// Oracle text:
//
//	Indestructible, haste
//	Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)
//	{1}, {T}: Target creature with power 2 or less can't be blocked this turn.
//	Hazoret can't attack or block unless you have max speed.
var HazoretGodseeker = &game.CardDef{
	Name: "Hazoret, Godseeker",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Red),
	}),
	Colors:        []mana.Color{mana.Red},
	ColorIdentity: mana.NewColorIdentity(mana.Red),
	Supertypes:    []types.Super{types.Legendary},
	Types:         []types.Card{types.Creature},
	Subtypes:      []types.Sub{types.Sub("God")},
	Power:         opt.Val(game.PT{Value: 5}),
	Toughness:     opt.Val(game.PT{Value: 3}),
	OracleText:    "Indestructible, haste\nStart your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)\n{1}, {T}: Target creature with power 2 or less can't be blocked this turn.\nHazoret can't attack or block unless you have max speed.",
	Abilities: []game.AbilityDef{
		game.IndestructibleAbility,
		game.HasteAbility,
		{
			Kind: game.TriggeredAbility,
			Text: "Start your engines!",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:  game.EventPermanentEnteredBattlefield,
					Source: game.TriggerSourceSelf,
				},
			}),
			Effects: []game.Effect{
				{Type: game.EffectStartEngines, TargetIndex: game.TargetIndexController},
			},
		},
		{
			Kind: game.ActivatedAbility,
			Text: "{1}, {T}: Target creature with power 2 or less can't be blocked this turn.",
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(1),
			}),
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature with power 2 or less",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Power:          opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2}),
					},
				},
			},
			Effects: []game.Effect{
				{
					Type:           game.EffectApplyRule,
					TargetIndex:    0,
					UntilEndOfTurn: true,
					RuleEffects: []game.RuleEffect{
						{Kind: game.RuleEffectCantBeBlocked},
					},
				},
			},
		},
		{
			Kind: game.StaticAbility,
			Text: "Hazoret can't attack or block unless you have max speed.",
			Condition: opt.Val(game.Condition{
				Text:                  "unless you have max speed",
				Negate:                true,
				ControllerHasMaxSpeed: true,
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
	},
}
