package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ReshapeTheEarth is the card definition for Reshape the Earth.
//
// Type: Sorcery
// Cost: {6}{G}{G}{G}
//
// Oracle text:
//
//	Search your library for up to ten land cards, put them onto the battlefield tapped, then shuffle.
var ReshapeTheEarth = newReshapeTheEarth

func newReshapeTheEarth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Reshape the Earth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
				cost.G,
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
								SourceZone:   zone.Library,
								Destination:  zone.Battlefield,
								Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}},
								EntersTapped: true,
							},
							Amount: game.Fixed(10),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to ten land cards, put them onto the battlefield tapped, then shuffle.
		`,
		},
	}
}
