package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlazemireVerge is the card definition for Blazemire Verge.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {B}.
//	{T}: Add {R}. Activate only if you control a Swamp or a Mountain.
var BlazemireVerge = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Red),
	CardFace: game.CardFace{
		Name:  "Blazemire Verge",
		Types: []types.Card{types.Land},
		OracleText: `
			{T}: Add {B}.
			{T}: Add {R}. Activate only if you control a Swamp or a Mountain.
		`,
		ManaAbilities: []game.ManaAbilityBody{
			{
				Text: `
					{T}: Add {B}.
				`,
				AdditionalCosts: []game.AdditionalCost{{Kind: game.AdditionalCostTap}},
				Content: game.PlainAbilityContent{
					Sequence: []game.Effect{
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.B, TargetIndex: game.TargetIndexController},
					},
				},
			},
			{
				Text: `
					{T}: Add {R}. Activate only if you control a Swamp or a Mountain.
				`,
				AdditionalCosts: []game.AdditionalCost{{Kind: game.AdditionalCostTap}},
				ActivationCondition: opt.Val(game.Condition{
					ControllerControls: game.PermanentFilter{
						SubtypesAny: []types.Sub{types.Swamp, types.Mountain},
					},
				}),
				Content: game.PlainAbilityContent{
					Sequence: []game.Effect{
						{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.R, TargetIndex: game.TargetIndexController},
					},
				},
			},
		},
	},
}
