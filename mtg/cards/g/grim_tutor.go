package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GrimTutor is the card definition for Grim Tutor.
//
// Type: Sorcery
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Search your library for a card, put that card into your hand, then shuffle. You lose 3 life.
var GrimTutor = newGrimTutor

func newGrimTutor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Grim Tutor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:       zone.Library,
								Destination:      zone.Hand,
								FailToFindPolicy: game.SearchMustFindIfAvailable,
							},
							Amount: game.Fixed(1),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a card, put that card into your hand, then shuffle. You lose 3 life.
		`,
		},
	}
}
