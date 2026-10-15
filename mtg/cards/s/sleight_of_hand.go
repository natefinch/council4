package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SleightOfHand is the card definition for Sleight of Hand.
//
// Type: Sorcery
// Cost: {U}
//
// Oracle text:
//
//	Look at the top two cards of your library. Put one of them into your hand and the other on the bottom of your library.
var SleightOfHand = newSleightOfHand

func newSleightOfHand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sleight of Hand",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(2),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderLibraryBottom,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top two cards of your library. Put one of them into your hand and the other on the bottom of your library.
		`,
		},
	}
}
