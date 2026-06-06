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
	CardFace: game.CardFace{
		Name:  "Command Tower",
		Types: []types.Card{types.Land},
		OracleText: `
			{T}: Add one mana of any color in your commander's color identity.
		`,
		ManaAbilities: []game.ManaAbilityBody{
			{
				Text: `
					{T}: Add one mana of any color in your commander's color identity.
				`,
				AdditionalCosts: []game.AdditionalCost{
					{Kind: game.AdditionalCostTap},
				},
				Content: game.PlainAbilityContent{
					Sequence: []game.Effect{
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
		},
	},
}
