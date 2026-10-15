package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TolarianWinds is the card definition for Tolarian Winds.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Discard all the cards in your hand, then draw that many cards.
var TolarianWinds = newTolarianWinds

func newTolarianWinds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tolarian Winds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Text: "Discard all the cards in your hand, then draw that many cards.",
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Player:     game.ControllerReference(),
							EntireHand: true,
						},
						PublishResult: game.ResultKey("discarded-this-way"),
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:      game.DynamicAmountPreviousEffectResult,
								ResultKey: game.ResultKey("discarded-this-way"),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Discard all the cards in your hand, then draw that many cards.
		`,
		},
	}
}
