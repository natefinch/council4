package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FaithfulMending is the card definition for Faithful Mending.
//
// Type: Instant
// Cost: {W}{U}
//
// Oracle text:
//
//	You gain 2 life, draw two cards, then discard two cards.
//	Flashback {1}{W}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var FaithfulMending = newFaithfulMending

func newFaithfulMending() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Faithful Mending",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Blue, color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {1}{W}{U}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.W,
								cost.U,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You gain 2 life, draw two cards, then discard two cards.
			Flashback {1}{W}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
