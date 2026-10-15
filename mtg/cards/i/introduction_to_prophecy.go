package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IntroductionToProphecy is the card definition for Introduction to Prophecy.
//
// Type: Sorcery — Lesson
// Cost: {3}
//
// Oracle text:
//
//	Scry 2, then draw a card.
var IntroductionToProphecy = newIntroductionToProphecy

func newIntroductionToProphecy() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Introduction to Prophecy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Lesson},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Scry{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Scry 2, then draw a card.
		`,
		},
	}
}
