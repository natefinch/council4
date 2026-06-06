package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MossfireValley is the card definition for Mossfire Valley.
//
// Type: Land
//
// Oracle text:
//
//	{1}, {T}: Add {R}{G}.
var MossfireValley = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green, color.Red),
	CardFace: game.CardFace{
		Name:  "Mossfire Valley",
		Types: []types.Card{types.Land},
		OracleText: `
			{1}, {T}: Add {R}{G}.
		`,
		ManaAbilities: []game.ManaAbilityBody{
			{
				Text: `
					{1}, {T}: Add {R}{G}.
				`,
				ManaCost: opt.Val(cost.Mana{
					cost.O(1),
				}),
				AdditionalCosts: []game.AdditionalCost{{Kind: game.AdditionalCostTap}},
				Content: game.PlainAbilityContent{
					Sequence: []game.Effect{
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.R, TargetIndex: game.TargetIndexController},
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.G, TargetIndex: game.TargetIndexController},
					},
				},
			},
		},
	},
}
