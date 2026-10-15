package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DigThroughTime is the card definition for Dig Through Time.
//
// Type: Instant
// Cost: {6}{U}{U}
//
// Oracle text:
//
//	Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
//	Look at the top seven cards of your library. Put two of them into your hand and the rest on the bottom of your library in any order.
var DigThroughTime = newDigThroughTime

func newDigThroughTime() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Dig Through Time",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.DelveStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(7),
							Take:      game.Fixed(2),
							Remainder: game.DigRemainderLibraryBottom,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
			Look at the top seven cards of your library. Put two of them into your hand and the rest on the bottom of your library in any order.
		`,
		},
	}
}
