package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EnvironmentalSciences is the card definition for Environmental Sciences.
//
// Type: Sorcery — Lesson
// Cost: {2}
//
// Oracle text:
//
//	Search your library for a basic land card, reveal it, put it into your hand, then shuffle. You gain 2 life.
var EnvironmentalSciences = newEnvironmentalSciences

func newEnvironmentalSciences() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Environmental Sciences",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Lesson},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
								Reveal:      true,
							},
							Amount: game.Fixed(1),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a basic land card, reveal it, put it into your hand, then shuffle. You gain 2 life.
		`,
		},
	}
}
