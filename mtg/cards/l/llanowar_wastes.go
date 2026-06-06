package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// LlanowarWastes is the card definition for Llanowar Wastes.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{T}: Add {B} or {G}. This land deals 1 damage to you.
var LlanowarWastes = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Green),
	CardFace: game.CardFace{
		Name:  "Llanowar Wastes",
		Types: []types.Card{types.Land},
		OracleText: `
			{T}: Add {C}.
			{T}: Add {B} or {G}. This land deals 1 damage to you.
		`,
		ManaAbilities: []game.ManaAbilityBody{
			{
				Text: `
					{T}: Add {C}.
				`,
				AdditionalCosts: []game.AdditionalCost{
					{
						Kind: game.AdditionalCostTap,
					},
				},
				Content: game.PlainAbilityContent{
					Sequence: []game.Instruction{
						{
							Primitive: game.AddMana{
								Amount:    game.Fixed(1),
								ManaColor: mana.C,
							},
						},
					},
				},
			},
			{
				Text: `
					{T}: Add {B} or {G}. This land deals 1 damage to you.
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
									Kind:   game.ResolutionChoiceMana,
									Prompt: "Choose {B} or {G}",
									Colors: []mana.Color{
										mana.B,
										mana.G,
									},
								},
								PublishChoice: game.ChoiceKey("llanowar-wastes-color"),
							},
						},
						{
							Primitive: game.AddMana{
								Amount:     game.Fixed(1),
								ChoiceFrom: game.ChoiceKey("llanowar-wastes-color"),
							},
						},
						{
							Primitive: game.Damage{
								Amount:    game.Fixed(1),
								Recipient: game.TargetRecipient(game.TargetIndexController),
							},
						},
					},
				},
			},
		},
	},
}
