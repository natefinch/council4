package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VampiricTutor is the card definition for Vampiric Tutor.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Search your library for a card, then shuffle and put that card on top. You lose 2 life.
var VampiricTutor = newVampiricTutor

func newVampiricTutor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vampiric Tutor",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
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
								FailToFindPolicy:    game.SearchMustFindIfAvailable,
							},
							Amount: game.Fixed(1),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a card, then shuffle and put that card on top. You lose 2 life.
		`,
		},
	}
}
