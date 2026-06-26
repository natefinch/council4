package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LilianaSSpoils is the card definition for Liliana's Spoils.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	Target opponent discards a card.
//	Look at the top five cards of your library. You may reveal a black card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var LilianaSSpoils = newLilianaSSpoils()

func newLilianaSSpoils() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Liliana's Spoils",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(5),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderLibraryBottom,
							Filter:    opt.Val(game.Selection{ColorsAny: []color.Color{color.Black}}),
							TakeUpTo:  true,
							Reveal:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent discards a card.
			Look at the top five cards of your library. You may reveal a black card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
