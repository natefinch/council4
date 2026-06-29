package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MysticMeditation is the card definition for Mystic Meditation.
//
// Type: Sorcery
// Cost: {3}{U}
//
// Oracle text:
//
//	Draw three cards. Then discard two cards unless you discard a creature card.
var MysticMeditation = newMysticMeditation()

func newMysticMeditation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mystic Meditation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Text: "Draw three cards. Then discard two cards unless you discard a creature card.",
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.DiscardUnlessType{
							Player:      game.ControllerReference(),
							Amount:      2,
							ExemptTypes: []types.Card{types.Creature},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw three cards. Then discard two cards unless you discard a creature card.
		`,
		},
	}
}
