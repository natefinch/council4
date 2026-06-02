package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Forest is the card definition for Forest.
//
// Type: Basic Land — Forest
//
// Oracle text:
//
//	({T}: Add {G}.)
var Forest = &game.CardDef{
	Name:          "Forest",
	ColorIdentity: color.NewIdentity(color.Green),
	Supertypes:    []types.Super{types.Basic},
	Types:         []types.Card{types.Land},
	Subtypes:      []types.Sub{types.Forest},
	OracleText:    "({T}: Add {G}.)",
	Abilities: []game.AbilityDef{
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add {G}.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{Type: game.EffectAddMana, Amount: 1, ManaColor: mana.G, TargetIndex: game.TargetIndexController},
			},
		},
	},
}
