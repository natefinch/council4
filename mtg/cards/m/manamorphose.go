package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Manamorphose is the card definition for Manamorphose.
//
// Type: Instant
// Cost: {1}{R/G}
//
// Oracle text:
//
//	Add two mana in any combination of colors.
//	Draw a card.
var Manamorphose = newManamorphose

func newManamorphose() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Manamorphose",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.R, mana.G),
			}),
			Colors: []color.Color{color.Green, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddMana{
							Amount:            game.Fixed(2),
							CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
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
			Add two mana in any combination of colors.
			Draw a card.
		`,
		},
	}
}
