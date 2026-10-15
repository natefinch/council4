package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ContactOtherPlane is the card definition for Contact Other Plane.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Roll a d20.
//	1—9 | Draw two cards.
//	10—19 | Scry 2, then draw two cards.
//	20 | Scry 3, then draw three cards.
var ContactOtherPlane = newContactOtherPlane

func newContactOtherPlane() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Contact Other Plane",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.RollDie{
							Sides: 20,
						},
						PublishResult: game.ResultKey("die-roll-result"),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "die-roll-result",
							AmountRange: opt.Val(game.IntRange{Min: 1, Max: 9}),
						}),
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "die-roll-result",
							AmountRange: opt.Val(game.IntRange{Min: 10, Max: 19}),
						}),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "die-roll-result",
							AmountRange: opt.Val(game.IntRange{Min: 10, Max: 19}),
						}),
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "die-roll-result",
							AmountRange: opt.Val(game.IntRange{Min: 20, Max: 20}),
						}),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "die-roll-result",
							AmountRange: opt.Val(game.IntRange{Min: 20, Max: 20}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Roll a d20.
			1—9 | Draw two cards.
			10—19 | Scry 2, then draw two cards.
			20 | Scry 3, then draw three cards.
		`,
		},
	}
}
