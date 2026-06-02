package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
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
	Name:          "Mossfire Valley",
	ColorIdentity: mana.NewColorIdentity(color.Green, color.Red),
	Types:         []types.Card{types.Land},
	OracleText:    "{1}, {T}: Add {R}{G}.",
	Abilities: []game.AbilityDef{
		{
			Kind:          game.ActivatedAbility,
			Text:          "{1}, {T}: Add {R}{G}.",
			IsManaAbility: true,
			ManaCost: opt.Val(mana.Cost{
				mana.GenericMana(1),
			}),
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{Type: game.EffectAddMana, Amount: 1, ManaColor: color.Red, TargetIndex: game.TargetIndexController},
				{Type: game.EffectAddMana, Amount: 1, ManaColor: color.Green, TargetIndex: game.TargetIndexController},
			},
		},
	},
}
