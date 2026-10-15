package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThreeVisits is the card definition for Three Visits.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Search your library for a Forest card, put it onto the battlefield, then shuffle.
var ThreeVisits = newThreeVisits

func newThreeVisits() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Three Visits",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Battlefield,
								Filter:      game.Selection{SubtypesAny: []types.Sub{types.Forest}},
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a Forest card, put it onto the battlefield, then shuffle.
		`,
		},
	}
}
