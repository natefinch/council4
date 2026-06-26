package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ForgingTheAnchor is the card definition for Forging the Anchor.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	Look at the top five cards of your library. You may reveal any number of artifact cards from among them and put the revealed cards into your hand. Put the rest on the bottom of your library in a random order.
var ForgingTheAnchor = newForgingTheAnchor()

func newForgingTheAnchor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Forging the Anchor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(5),
							Take:      game.Fixed(5),
							Remainder: game.DigRemainderLibraryBottom,
							Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
							TakeUpTo:  true,
							Reveal:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top five cards of your library. You may reveal any number of artifact cards from among them and put the revealed cards into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
