package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VisionOfLove is the card definition for Vision of Love.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	You may sacrifice an artifact or discard a card. If you do, draw two cards.
var VisionOfLove = newVisionOfLove()

func newVisionOfLove() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Vision of Love",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Amount:    game.Fixed(1),
							Player:    game.ControllerReference(),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
						},
						Optional:      true,
						PublishResult: game.ResultKey("disjunctive-cost-a"),
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:      "disjunctive-cost-a",
							Accepted: game.TriFalse,
						}),
						Optional:      true,
						PublishResult: game.ResultKey("disjunctive-cost-b"),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "disjunctive-cost-a",
							Succeeded: game.TriTrue,
						}),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "disjunctive-cost-b",
							Succeeded: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			You may sacrifice an artifact or discard a card. If you do, draw two cards.
		`,
		},
	}
}
