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
	ColorIdentity: color.NewIdentity(color.Green, color.Red),
	CardFace: game.CardFace{
		Name:  "Karplusan Forest",
		Types: []types.Card{types.Land},
		OracleText: `
			{T}: Add {C}.
			{T}: Add {R} or {G}. This land deals 1 damage to you.
		`,
		ManaAbilities: []game.ManaAbilityBody{
			{
				Text: `
					{T}: Add {C}.
				`,
				AdditionalCosts: []game.AdditionalCost{{Kind: game.AdditionalCostTap}},
				Content: game.PlainAbilityContent{
					Sequence: []game.Effect{
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.C, TargetIndex: game.TargetIndexController},
					},
				},
			},
			{
				Text: `
					{T}: Add {R} or {G}. This land deals 1 damage to you.
				`,
				AdditionalCosts: []game.AdditionalCost{{Kind: game.AdditionalCostTap}},
				Content: game.PlainAbilityContent{
					Sequence: []game.Effect{
						{
							Type:        game.EffectChoose,
							TargetIndex: game.TargetIndexController,
							Choice: opt.Val(game.ResolutionChoice{
								Kind:   game.ResolutionChoiceMana,
								Prompt: "Choose {R} or {G}",
								Colors: []mana.Color{mana.R, mana.G},
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
		},
	},
}
