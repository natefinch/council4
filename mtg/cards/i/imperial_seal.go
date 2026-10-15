package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ImperialSeal is the card definition for Imperial Seal.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Search your library for a card, then shuffle and put that card on top. You lose 2 life.
var ImperialSeal = newImperialSeal

func newImperialSeal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Imperial Seal",
			ManaCost: opt.Val(cost.Mana{
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
