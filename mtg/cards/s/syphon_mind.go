package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SyphonMind is the card definition for Syphon Mind.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	Each other player discards a card. You draw a card for each card discarded this way.
var SyphonMind = newSyphonMind

func newSyphonMind() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Syphon Mind",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount:      game.Fixed(1),
							PlayerGroup: game.OpponentsReference(),
						},
						PublishResult: game.ResultKey("actual-card-quantity-1"),
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("actual-card-quantity-1")},
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 1,
								Selection:  &game.Selection{},
								ResultKey:  game.ResultKey("actual-card-quantity-1"),
							}),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:      "actual-card-quantity-1",
							Accepted: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Each other player discards a card. You draw a card for each card discarded this way.
		`,
		},
	}
}
