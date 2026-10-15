package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ContentiousPlan is the card definition for Contentious Plan.
//
// Type: Sorcery
// Cost: {1}{U}
//
// Oracle text:
//
//	Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
//	Draw a card.
var ContentiousPlan = newContentiousPlan

func newContentiousPlan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Contentious Plan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Proliferate{
							Amount: game.Fixed(1),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
			Draw a card.
		`,
		},
	}
}
