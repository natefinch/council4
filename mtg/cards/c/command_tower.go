package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
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
					{
						Kind: game.AdditionalCostTap,
					},
				},
				Content: game.PlainAbilityContent{
					Sequence: []game.Instruction{
						{
							Primitive: game.Choose{
								Choice: game.ResolutionChoice{
									Kind:        game.ResolutionChoiceMana,
									Prompt:      "Choose a color in your commander's color identity",
									ColorSource: game.ResolutionChoiceColorSourceCommanderIdentity,
								},
								PublishChoice: game.ChoiceKey("command-tower-color"),
							},
						},
						{
							Primitive: game.AddMana{
								Amount:     game.Fixed(1),
								ChoiceFrom: game.ChoiceKey("command-tower-color"),
							},
						},
					},
				},
			},
		},
	},
}
