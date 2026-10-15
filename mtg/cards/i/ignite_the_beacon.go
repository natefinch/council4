package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IgniteTheBeacon is the card definition for Ignite the Beacon.
//
// Type: Instant
// Cost: {4}{W}
//
// Oracle text:
//
//	Search your library for up to two planeswalker cards, reveal them, put them into your hand, then shuffle.
var IgniteTheBeacon = newIgniteTheBeacon

func newIgniteTheBeacon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ignite the Beacon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
								Reveal:      true,
							},
							Amount: game.Fixed(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to two planeswalker cards, reveal them, put them into your hand, then shuffle.
		`,
		},
	}
}
