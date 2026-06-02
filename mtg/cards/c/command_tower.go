package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CommandTower is the card definition for Command Tower.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add one mana of any color in your commander's color identity.
var CommandTower = &game.CardDef{
	Name:       "Command Tower",
	Types:      []types.Card{types.Land},
	OracleText: "{T}: Add one mana of any color in your commander's color identity.",
	Abilities: []game.AbilityDef{
		{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add one mana of any color in your commander's color identity.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectChoose,
					TargetIndex: game.TargetIndexController,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:        game.ResolutionChoiceMana,
						Prompt:      "Choose a color in your commander's color identity",
						ColorSource: game.ResolutionChoiceColorSourceCommanderIdentity,
					}),
					LinkID: "command-tower-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					TargetIndex:  game.TargetIndexController,
					ChoiceLinkID: "command-tower-color",
				},
			},
		},
	},
}
