// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTLibraryModes is the card definition for RT Library Modes.
//
// Type: Sorcery
//
// Oracle text:
//
//	Choose one —
//	• Reveal the top card of your library.
//	• You may look at the top card of your library.
var RTLibraryModes = newRTLibraryModes

func newRTLibraryModes() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Library Modes",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Reveal the top card of your library.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Reveal{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "You may look at the top card of your library.",
						Sequence: []game.Instruction{
							{
								Primitive: game.LookAtLibraryTop{
									Player:        game.ControllerReference(),
									PublishLinked: game.LinkedKey("sequence-effect-0-product"),
								},
								Optional: true,
								LocalProducts: game.LocalProducts{
									Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Reveal the top card of your library.
			• You may look at the top card of your library.
		`,
		},
	}
}
