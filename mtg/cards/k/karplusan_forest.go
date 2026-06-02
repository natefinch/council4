package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KarplusanForest is the card definition for Karplusan Forest.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{T}: Add {R} or {G}. This land deals 1 damage to you.
var KarplusanForest = &game.CardDef{
	Name:          "Karplusan Forest",
	ColorIdentity: mana.NewColorIdentity(color.Green, color.Red),
	Types:         []types.Card{types.Land},
	OracleText:    "{T}: Add {C}.\n{T}: Add {R} or {G}. This land deals 1 damage to you.",
	Abilities: []game.AbilityDef{
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {C}.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{Type: game.EffectAddMana, Amount: 1, ManaColor: color.Colorless, TargetIndex: game.TargetIndexController},
			},
		},
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {R} or {G}. This land deals 1 damage to you.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectChoose,
					TargetIndex: game.TargetIndexController,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:   game.ResolutionChoiceColor,
						Prompt: "Choose {R} or {G}",
						Colors: []color.Color{color.Red, color.Green},
					}),
					LinkID: "karplusan-forest-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  game.TargetIndexController,
					ChoiceLinkID: "karplusan-forest-color",
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
