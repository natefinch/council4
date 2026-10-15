package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DrasticRevelation is the card definition for Drastic Revelation.
//
// Type: Sorcery
// Cost: {2}{U}{B}{R}
//
// Oracle text:
//
//	Discard your hand. Draw seven cards, then discard three cards at random.
var DrasticRevelation = newDrasticRevelation

func newDrasticRevelation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Drastic Revelation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.B,
				cost.R,
			}),
			Colors: []color.Color{color.Black, color.Red, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Player:     game.ControllerReference(),
							EntireHand: true,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(7),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Discard{
							Amount:   game.Fixed(3),
							Player:   game.ControllerReference(),
							AtRandom: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Discard your hand. Draw seven cards, then discard three cards at random.
		`,
		},
	}
}
