package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BridgeworksBattle is the card definition for Bridgeworks Battle // Tanglespan Bridgeworks.
//
// Type: Sorcery // Land
// Face: Bridgeworks Battle — Sorcery ({2}{G})
// Face: Tanglespan Bridgeworks — Land
//
// Front oracle text:
//
//	Target creature you control gets +2/+2 until end of turn. It fights up to
//	one target creature you don't control. (Each deals damage equal to its power
//	to the other.)
//
// Back oracle text:
//
//	As this land enters, you may pay 3 life. If you don't, it enters tapped.
//	{T}: Add {G}.
var BridgeworksBattle = &game.CardDef{
	Name: "Bridgeworks Battle",
	ManaCost: opt.Val(cost.Mana{
		cost.O(2),
		cost.G,
	}),
	Colors:        []color.Color{color.Green},
	ColorIdentity: color.NewIdentity(color.Green),
	Types:         []types.Card{types.Sorcery},
	OracleText:    "Target creature you control gets +2/+2 until end of turn. It fights up to one target creature you don't control. (Each deals damage equal to its power to the other.)",
	Layout:        game.LayoutModalDFC,
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Target creature you control gets +2/+2 until end of turn. It fights up to one target creature you don't control.",
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
					// "up to one" — may choose zero or one
					MinTargets: 0,
					MaxTargets: 1,
					Constraint: "creature you don't control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerNotYou,
					},
				},
			},
			Effects: []game.Effect{
				{
					Type:           game.EffectModifyPT,
					PowerDelta:     2,
					ToughnessDelta: 2,
					TargetIndex:    0,
					UntilEndOfTurn: true,
				},
				{
					Type:               game.EffectFight,
					TargetIndex:        0,
					RelatedTargetIndex: opt.Val(1),
					Description:        "target creature you control fights up to one target creature you don't control",
				},
			},
		},
	},
	Back: opt.Val(game.CardFace{
		Name:  "Tanglespan Bridgeworks",
		Types: []types.Card{types.Land},
		EntersTappedUnlessPaid: opt.Val(game.ResolutionPayment{
			Prompt: "Pay 3 life?",
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostPayLife, Amount: 3, Text: "Pay 3 life"},
			},
		}),
		OracleText: "As this land enters, you may pay 3 life. If you don't, it enters tapped.\n{T}: Add {G}.",
		Abilities: []game.AbilityDef{
			{
				Kind:          game.ActivatedAbility,
				Text:          "{T}: Add {G}.",
				IsManaAbility: true,
				AdditionalCosts: []game.AdditionalCost{
					{Kind: game.AdditionalCostTap},
				},
				Effects: []game.Effect{
					{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.G, TargetIndex: game.TargetIndexController},
				},
			},
		},
	}),
}
