package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Deduce is the card definition for Deduce.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Draw a card. Investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
var Deduce = newDeduce

func newDeduce() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Deduce",
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
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Investigate{
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw a card. Investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
		`,
		},
	}
}
