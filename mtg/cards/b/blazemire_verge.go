package b

import (
	"github.com/natefinch/council4/mtg/game"
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
	Name:          "Blazemire Verge",
	ColorIdentity: mana.NewColorIdentity(mana.Black, mana.Red),
	Types:         []types.Card{types.Land},
	OracleText:    "{T}: Add {B}.\n{T}: Add {R}. Activate only if you control a Swamp or a Mountain.",
	Abilities: []game.AbilityDef{
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {B}.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.Black, TargetIndex: game.TargetIndexController},
			},
		},
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {R}. Activate only if you control a Swamp or a Mountain.",
			IsManaAbility: true,
			ActivationCondition: opt.Val(game.Condition{
				ControllerControls: game.PermanentFilter{
					SubtypesAny: []types.Sub{types.Swamp, types.Mountain},
				},
			}),
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.Red, TargetIndex: game.TargetIndexController},
			},
		},
	},
}
