package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DecayingTimeLoop is the card definition for Decaying Time Loop.
//
// Type: Instant
// Cost: {3}{R}
//
// Oracle text:
//
//	Discard all the cards in your hand, then draw that many cards.
//	Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
var DecayingTimeLoop = newDecayingTimeLoop

func newDecayingTimeLoop() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Decaying Time Loop",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.RetraceStaticBody,
			},
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
			Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
		`,
		},
	}
}
