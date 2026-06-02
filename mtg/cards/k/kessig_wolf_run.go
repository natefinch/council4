package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KessigWolfRun is the card definition for Kessig Wolf Run.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn.
var KessigWolfRun = &game.CardDef{
	Name:          "Kessig Wolf Run",
	ColorIdentity: mana.NewColorIdentity(color.Green, color.Red),
	Types:         []types.Card{types.Land},
	OracleText:    "{T}: Add {C}.\n{X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn.",
	Abilities: []game.AbilityDef{
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {C}.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.C, TargetIndex: game.TargetIndexController},
			},
		},
		{
			Kind: game.ActivatedAbility,
			Text: "{X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn.",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.G,
			}),
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
					},
				},
			},
			Effects: []game.Effect{
				{
					Type:           game.EffectModifyPT,
					TargetIndex:    0,
					UntilEndOfTurn: true,
					PowerDeltaDynamic: opt.Val(game.DynamicAmount{
						Kind: game.DynamicAmountX,
					}),
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
