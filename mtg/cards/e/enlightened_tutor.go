package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EnlightenedTutor is the card definition for Enlightened Tutor.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Search your library for an artifact or enchantment card, reveal it, then shuffle and put that card on top.
var EnlightenedTutor = newEnlightenedTutor

func newEnlightenedTutor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Enlightened Tutor",
			ManaCost: opt.Val(cost.Mana{
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
								SourceZone:          zone.Library,
								Destination:         zone.Library,
								DestinationPosition: game.SearchPositionTop,
								Filter:              game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}},
								Reveal:              true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for an artifact or enchantment card, reveal it, then shuffle and put that card on top.
		`,
		},
	}
}
