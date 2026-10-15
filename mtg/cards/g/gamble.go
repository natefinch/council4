package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Gamble is the card definition for Gamble.
//
// Type: Sorcery
// Cost: {R}
//
// Oracle text:
//
//	Search your library for a card, put that card into your hand, discard a card at random, then shuffle.
var Gamble = newGamble

func newGamble() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Gamble",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
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
						Primitive: game.Discard{
							Amount:   game.Fixed(1),
							Player:   game.ControllerReference(),
							AtRandom: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a card, put that card into your hand, discard a card at random, then shuffle.
		`,
		},
	}
}
