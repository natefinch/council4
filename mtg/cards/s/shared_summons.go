package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SharedSummons is the card definition for Shared Summons.
//
// Type: Instant
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Search your library for up to two creature cards with different names, reveal them, put them into your hand, then shuffle.
var SharedSummons = newSharedSummons

func newSharedSummons() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Shared Summons",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:     zone.Library,
								Destination:    zone.Hand,
								Filter:         game.Selection{RequiredTypes: []types.Card{types.Creature}},
								Reveal:         true,
								DifferentNames: true,
							},
							Amount: game.Fixed(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to two creature cards with different names, reveal them, put them into your hand, then shuffle.
		`,
		},
	}
}
