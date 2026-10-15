package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WordsOfWisdom is the card definition for Words of Wisdom.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	You draw two cards, then each other player draws a card.
var WordsOfWisdom = newWordsOfWisdom

func newWordsOfWisdom() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Words of Wisdom",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Draw{
							Amount:      game.Fixed(1),
							PlayerGroup: game.OpponentsReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You draw two cards, then each other player draws a card.
		`,
		},
	}
}
