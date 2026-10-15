package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TrapmakerSSnare is the card definition for Trapmaker's Snare.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Search your library for a Trap card, reveal it, put it into your hand, then shuffle.
var TrapmakerSSnare = newTrapmakerSSnare

func newTrapmakerSSnare() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Trapmaker's Snare",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{SubtypesAny: []types.Sub{types.Trap}},
								Reveal:      true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a Trap card, reveal it, put it into your hand, then shuffle.
		`,
		},
	}
}
