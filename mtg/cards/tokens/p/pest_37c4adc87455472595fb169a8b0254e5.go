package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Pest
//
// Type: Token Creature — Pest
//
// Oracle text:
//   When this creature dies, you gain 1 life.

// PestToken37c4adc87455472595fb169a8b0254e5 is the card definition for Pest.
var PestToken37c4adc87455472595fb169a8b0254e5 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Green),
	CardFace: game.CardFace{
		Name:      "Pest",
		Colors:    []color.Color{color.Black, color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Pest},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		TriggeredAbilities: []game.TriggeredAbility{
			game.TriggeredAbility{
				Text: "When this creature dies, you gain 1 life.",
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentDied,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{
					Sequence: []game.Instruction{
						{
							Primitive: game.GainLife{
								Amount: game.Fixed(1),
								Player: game.ControllerReference(),
							},
						},
					},
				}.Ability(),
			},
		},
		OracleText: `
			When this creature dies, you gain 1 life.
		`,
	},
}
