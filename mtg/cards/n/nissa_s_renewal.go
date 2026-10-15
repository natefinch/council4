package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NissaSRenewal is the card definition for Nissa's Renewal.
//
// Type: Sorcery
// Cost: {5}{G}
//
// Oracle text:
//
//	Search your library for up to three basic land cards, put them onto the battlefield tapped, then shuffle. You gain 7 life.
var NissaSRenewal = newNissaSRenewal

func newNissaSRenewal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nissa's Renewal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
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
							Amount: game.Fixed(3),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(7),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for up to three basic land cards, put them onto the battlefield tapped, then shuffle. You gain 7 life.
		`,
		},
	}
}
