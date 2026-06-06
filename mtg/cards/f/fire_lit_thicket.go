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
	ColorIdentity: color.NewIdentity(color.Green, color.Red),
	CardFace: game.CardFace{
		Name:  "Fire-Lit Thicket",
		Types: []types.Card{types.Land},
		OracleText: `
			{T}: Add {C}.
			{R/G}, {T}: Add {R}{R}, {R}{G}, or {G}{G}.
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
					{R/G}, {T}: Add {R}{R}, {R}{G}, or {G}{G}.
				`,
				ManaCost: opt.Val(cost.Mana{
					cost.HybridMana(mana.R, mana.G),
				}),
				AdditionalCosts: []game.AdditionalCost{{Kind: game.AdditionalCostTap}},
				Content: game.ModalAbilityContent{
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
		},
	},
}
