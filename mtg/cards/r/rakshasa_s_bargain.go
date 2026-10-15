package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RakshasaSBargain is the card definition for Rakshasa's Bargain.
//
// Type: Instant
// Cost: {2/B}{2/G}{2/U}
//
// Oracle text:
//
//	Look at the top four cards of your library. Put two of them into your hand and the rest into your graveyard.
var RakshasaSBargain = newRakshasaSBargain

func newRakshasaSBargain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Rakshasa's Bargain",
			ManaCost: opt.Val(cost.Mana{
				cost.Twobrid(mana.B),
				cost.Twobrid(mana.G),
				cost.Twobrid(mana.U),
			}),
			Colors: []color.Color{color.Black, color.Green, color.Blue},
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
			Look at the top four cards of your library. Put two of them into your hand and the rest into your graveyard.
		`,
		},
	}
}
