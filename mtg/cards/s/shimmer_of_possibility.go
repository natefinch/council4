package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShimmerOfPossibility is the card definition for Shimmer of Possibility.
//
// Type: Sorcery
// Cost: {1}{U}
//
// Oracle text:
//
//	Look at the top four cards of your library. Put one of them into your hand and the rest on the bottom of your library in a random order.
var ShimmerOfPossibility = newShimmerOfPossibility

func newShimmerOfPossibility() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Shimmer of Possibility",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(4),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderLibraryBottom,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top four cards of your library. Put one of them into your hand and the rest on the bottom of your library in a random order.
		`,
		},
	}
}
