package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LlanowarWastes is the card definition for Llanowar Wastes.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{T}: Add {B} or {G}. This land deals 1 damage to you.
var LlanowarWastes = &game.CardDef{
	Name:          "Llanowar Wastes",
	ColorIdentity: mana.NewColorIdentity(color.Black, color.Green),
	Types:         []types.Card{types.Land},
	OracleText:    "{T}: Add {C}.\n{T}: Add {B} or {G}. This land deals 1 damage to you.",
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
			Text:          "{T}: Add {B} or {G}. This land deals 1 damage to you.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectChoose,
					TargetIndex: game.TargetIndexController,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:   game.ResolutionChoiceMana,
						Prompt: "Choose {B} or {G}",
						Colors: []mana.Color{mana.B, mana.G},
					}),
					LinkID: "llanowar-wastes-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  game.TargetIndexController,
					ChoiceLinkID: "llanowar-wastes-color",
				},
				{
					Type:        game.EffectDamage,
					Amount:      1,
					TargetIndex: game.TargetIndexController,
				},
			},
		},
	},
}
