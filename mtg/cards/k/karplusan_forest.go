package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// KarplusanForest is the card definition for Karplusan Forest.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{T}: Add {R} or {G}. This land deals 1 damage to you.
var KarplusanForest = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green, color.Red),
	CardFace: game.CardFace{
		Name:  "Karplusan Forest",
		Types: []types.Card{types.Land},
		OracleText: `
			{T}: Add {C}.
			{T}: Add {R} or {G}. This land deals 1 damage to you.
		`,
		ManaAbilities: []game.ManaAbilityBody{
			{
				Text: `
					{T}: Add {C}.
				`,
				AdditionalCosts: []cost.Additional{
					{
						Kind: cost.AdditionalTap,
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
					{T}: Add {R} or {G}. This land deals 1 damage to you.
				`,
				AdditionalCosts: []cost.Additional{
					{
						Kind: cost.AdditionalTap,
					},
				},
				Content: game.PlainAbilityContent{
					Sequence: []game.Instruction{
						{
							Primitive: game.Choose{
								Choice: game.ResolutionChoice{
									Kind:   game.ResolutionChoiceMana,
									Prompt: "Choose {R} or {G}",
									Colors: []mana.Color{
										mana.R,
										mana.G,
									},
								},
								PublishChoice: game.ChoiceKey("karplusan-forest-color"),
							},
						},
						{
							Primitive: game.AddMana{
								Amount:     game.Fixed(1),
								ChoiceFrom: game.ChoiceKey("karplusan-forest-color"),
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
