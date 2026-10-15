package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThreeDreams is the card definition for Three Dreams.
//
// Type: Sorcery
// Cost: {4}{W}
//
// Oracle text:
//
//	Search your library for up to three Aura cards with different names, reveal them, put them into your hand, then shuffle.
var ThreeDreams = newThreeDreams

func newThreeDreams() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Three Dreams",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:     zone.Library,
								Destination:    zone.Hand,
								Filter:         game.Selection{SubtypesAny: []types.Sub{types.Aura}},
								Reveal:         true,
								DifferentNames: true,
							},
							Amount: game.Fixed(3),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to three Aura cards with different names, reveal them, put them into your hand, then shuffle.
		`,
		},
	}
}
