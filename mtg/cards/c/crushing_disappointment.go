package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrushingDisappointment is the card definition for Crushing Disappointment.
//
// Type: Instant
// Cost: {3}{B}
//
// Oracle text:
//
//	Each player loses 2 life. You draw two cards.
var CrushingDisappointment = newCrushingDisappointment

func newCrushingDisappointment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Crushing Disappointment",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.LoseLife{
							Amount:      game.Fixed(2),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player loses 2 life. You draw two cards.
		`,
		},
	}
}
