package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CycleOfRenewal is the card definition for Cycle of Renewal.
//
// Type: Instant — Lesson
// Cost: {2}{G}
//
// Oracle text:
//
//	Sacrifice a land. Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.
var CycleOfRenewal = newCycleOfRenewal

func newCycleOfRenewal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cycle of Renewal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Lesson},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Player:    game.ControllerReference(),
							Amount:    game.Fixed(1),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:   zone.Library,
								Destination:  zone.Battlefield,
								Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
								EntersTapped: true,
							},
							Amount: game.Fixed(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Sacrifice a land. Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.
		`,
		},
	}
}
