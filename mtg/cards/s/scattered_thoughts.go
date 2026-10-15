package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScatteredThoughts is the card definition for Scattered Thoughts.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Look at the top four cards of your library. Put two of those cards into your hand and the rest into your graveyard.
var ScatteredThoughts = newScatteredThoughts

func newScatteredThoughts() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Scattered Thoughts",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player: game.ControllerReference(),
							Look:   game.Fixed(4),
							Take:   game.Fixed(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top four cards of your library. Put two of those cards into your hand and the rest into your graveyard.
		`,
		},
	}
}
