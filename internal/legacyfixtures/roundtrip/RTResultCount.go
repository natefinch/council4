// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTResultCount is the card definition for RT Result Count.
//
// Type: Sorcery
//
// Oracle text:
//
//	Discard three cards. If you discarded two land cards this way, draw a card. Otherwise, you gain 2 life.
var RTResultCount = newRTResultCount

func newRTResultCount() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Result Count",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
						PublishResult: game.ResultKey("if-you-do"),
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("if-you-do")},
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:              "if-you-do",
							Succeeded:        game.TriTrue,
							ObjectSelection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
							ObjectCountRange: opt.Val(game.IntRange{Min: 2, Max: 2}),
							CardOnly:         true,
						}),
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:              "if-you-do",
							Succeeded:        game.TriTrue,
							ObjectSelection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
							ObjectCountRange: opt.Val(game.IntRange{Min: 2, Max: 2}),
							Negate:           true,
							CardOnly:         true,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Discard three cards. If you discarded two land cards this way, draw a card. Otherwise, you gain 2 life.
		`,
		},
	}
}
