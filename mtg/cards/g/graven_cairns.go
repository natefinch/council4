package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GravenCairns is the card definition for Graven Cairns.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{B/R}, {T}: Add {B}{B}, {B}{R}, or {R}{R}.
//
// The second ability is modeled as two independent color choices from {B, R},
// which covers the three legal outputs: {B}{B}, {B}{R}, and {R}{R}.
var GravenCairns = &game.CardDef{
	Name:          "Graven Cairns",
	ColorIdentity: color.NewIdentity(color.Black, color.Red),
	Types:         []types.Card{types.Land},
	OracleText:    "{T}: Add {C}.\n{B/R}, {T}: Add {B}{B}, {B}{R}, or {R}{R}.",
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
			Kind:          game.ActivatedAbility,
			Text:          "{B/R}, {T}: Add {B}{B}, {B}{R}, or {R}{R}.",
			IsManaAbility: true,
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.R),
			}),
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectChoose,
					TargetIndex: game.TargetIndexController,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:   game.ResolutionChoiceMana,
						Prompt: "Choose first mana color ({B} or {R})",
						Colors: []mana.Color{mana.B, mana.R},
					}),
					LinkID: "graven-cairns-color-1",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  game.TargetIndexController,
					ChoiceLinkID: "graven-cairns-color-1",
				},
				{
					Type:        game.EffectChoose,
					TargetIndex: game.TargetIndexController,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:   game.ResolutionChoiceMana,
						Prompt: "Choose second mana color ({B} or {R})",
						Colors: []mana.Color{mana.B, mana.R},
					}),
					LinkID: "graven-cairns-color-2",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  game.TargetIndexController,
					ChoiceLinkID: "graven-cairns-color-2",
				},
			},
		},
	},
}
