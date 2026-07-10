package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MapTheWastes is the card definition for Map the Wastes.
//
// Type: Sorcery
// Cost: {2}{G}
//
// Oracle text:
//
//	Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
var MapTheWastes = newMapTheWastes

func newMapTheWastes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Map the Wastes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
								SourceZone:   zone.Library,
								Destination:  zone.Battlefield,
								Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
								EntersTapped: true,
							},
							Amount: game.Fixed(1),
						},
					},
					{
						Primitive: game.Bolster{
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
		`,
		},
	}
}
