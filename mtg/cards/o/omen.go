package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Omen is the card definition for Omen.
//
// Type: Sorcery
// Cost: {1}{U}
//
// Oracle text:
//
//	Look at the top three cards of your library, then put them back in any order. You may shuffle.
//	Draw a card.
var Omen = newOmen

func newOmen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Omen",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ReorderLibraryTop{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.ShuffleLibrary{
							Player: game.ControllerReference(),
						},
						Optional: true,
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
			Look at the top three cards of your library, then put them back in any order. You may shuffle.
			Draw a card.
		`,
		},
	}
}
