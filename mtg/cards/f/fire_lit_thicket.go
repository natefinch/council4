package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FireLitThicket is the card definition for Fire-Lit Thicket.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{R/G}, {T}: Add {R}{R}, {R}{G}, or {G}{G}.
var FireLitThicket = &game.CardDef{
	Name:          "Fire-Lit Thicket",
	ColorIdentity: color.NewIdentity(color.Green, color.Red),
	Types:         []types.Card{types.Land},
	OracleText:    "{T}: Add {C}.\n{R/G}, {T}: Add {R}{R}, {R}{G}, or {G}{G}.",
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
			Text:          "{R/G}, {T}: Add {R}{R}, {R}{G}, or {G}{G}.",
			IsManaAbility: true,
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.R, mana.G),
			}),
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Modes: []game.Mode{
				{
					Text: "Add {R}{R}.",
					Effects: []game.Effect{
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.R, TargetIndex: game.TargetIndexController},
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.R, TargetIndex: game.TargetIndexController},
					},
				},
				{
					Text: "Add {R}{G}.",
					Effects: []game.Effect{
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.R, TargetIndex: game.TargetIndexController},
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.G, TargetIndex: game.TargetIndexController},
					},
				},
				{
					Text: "Add {G}{G}.",
					Effects: []game.Effect{
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.G, TargetIndex: game.TargetIndexController},
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.G, TargetIndex: game.TargetIndexController},
					},
				},
			},
		},
	},
}
