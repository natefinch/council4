package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WorldlyTutor is the card definition for Worldly Tutor.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Search your library for a creature card, reveal it, then shuffle and put the card on top.
var WorldlyTutor = newWorldlyTutor

func newWorldlyTutor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Worldly Tutor",
			ManaCost: opt.Val(cost.Mana{
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
								SourceZone:          zone.Library,
								Destination:         zone.Library,
								DestinationPosition: game.SearchPositionTop,
								Filter:              game.Selection{RequiredTypes: []types.Card{types.Creature}},
								Reveal:              true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a creature card, reveal it, then shuffle and put the card on top.
		`,
		},
	}
}
