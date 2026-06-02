package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Mountain is the card definition for Mountain.
//
// Type: Basic Land — Mountain
//
// Oracle text:
//
//	({T}: Add {R}.)
var Mountain = &game.CardDef{
	Name:          "Mountain",
	ColorIdentity: color.NewIdentity(color.Red),
	Supertypes:    []types.Super{types.Basic},
	Types:         []types.Card{types.Land},
	Subtypes:      []types.Sub{types.Mountain},
	OracleText:    "({T}: Add {R}.)",
	Abilities: []game.AbilityDef{
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {R}.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.R, TargetIndex: game.TargetIndexController},
			},
		},
	},
}
