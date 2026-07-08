package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PeerThroughDepths is the card definition for Peer Through Depths.
//
// Type: Instant — Arcane
// Cost: {1}{U}
//
// Oracle text:
//
//	Look at the top five cards of your library. You may reveal an instant or sorcery card from among them and put it into your hand. Put the rest on the bottom of your library in any order.
var PeerThroughDepths = newPeerThroughDepths

func newPeerThroughDepths() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Peer Through Depths",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Arcane},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(5),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderLibraryBottom,
							Filter:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}}),
							TakeUpTo:  true,
							Reveal:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top five cards of your library. You may reveal an instant or sorcery card from among them and put it into your hand. Put the rest on the bottom of your library in any order.
		`,
		},
	}
}
