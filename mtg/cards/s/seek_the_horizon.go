package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SeekTheHorizon is the card definition for Seek the Horizon.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Search your library for up to three basic land cards, reveal them, put them into your hand, then shuffle.
var SeekTheHorizon = newSeekTheHorizon

func newSeekTheHorizon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Seek the Horizon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
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
								Destination: zone.Hand,
								Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
								Reveal:      true,
							},
							Amount: game.Fixed(3),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to three basic land cards, reveal them, put them into your hand, then shuffle.
		`,
		},
	}
}
