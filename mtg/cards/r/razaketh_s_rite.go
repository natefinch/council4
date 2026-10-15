package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RazakethSRite is the card definition for Razaketh's Rite.
//
// Type: Sorcery
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Search your library for a card, put that card into your hand, then shuffle.
//	Cycling {B} ({B}, Discard this card: Draw a card.)
var RazakethSRite = newRazakethSRite

func newRazakethSRite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Razaketh's Rite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.B}),
			},
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
				},
			}.Ability()),
			OracleText: `
			Search your library for a card, put that card into your hand, then shuffle.
			Cycling {B} ({B}, Discard this card: Draw a card.)
		`,
		},
	}
}
