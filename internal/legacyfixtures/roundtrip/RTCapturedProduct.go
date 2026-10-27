// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTCapturedProduct is the card definition for RT Captured Product.
//
// Type: Sorcery
//
// Oracle text:
//
//	Create a 1/1 green Insect creature token. You gain 1 life. Sacrifice it at the beginning of the next end step.
var RTCapturedProduct = newRTCapturedProduct

func newRTCapturedProduct() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Captured Product",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:        game.Fixed(1),
							Source:        game.TokenDef(rTCapturedProductToken),
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Sacrifice{
														Group: game.CapturedObjectsGroup(),
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								CapturedObjectGroup: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 1/1 green Insect creature token. You gain 1 life. Sacrifice it at the beginning of the next end step.
		`,
		},
	}
}

var rTCapturedProductToken = newRTCapturedProductToken()

func newRTCapturedProductToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Insect",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
