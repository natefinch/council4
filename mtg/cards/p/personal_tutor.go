package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PersonalTutor is the card definition for Personal Tutor.
//
// Type: Sorcery
// Cost: {U}
//
// Oracle text:
//
//	Search your library for a sorcery card, reveal it, then shuffle and put that card on top.
var PersonalTutor = newPersonalTutor

func newPersonalTutor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Personal Tutor",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:          zone.Library,
								Destination:         zone.Library,
								DestinationPosition: game.SearchPositionTop,
								Filter:              game.Selection{RequiredTypes: []types.Card{types.Sorcery}},
								Reveal:              true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a sorcery card, reveal it, then shuffle and put that card on top.
		`,
		},
	}
}
