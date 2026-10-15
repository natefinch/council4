package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ControlOfTheCourt is the card definition for Control of the Court.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Draw four cards, then discard three cards at random.
var ControlOfTheCourt = newControlOfTheCourt

func newControlOfTheCourt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Control of the Court",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(4),
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
			Draw four cards, then discard three cards at random.
		`,
		},
	}
}
