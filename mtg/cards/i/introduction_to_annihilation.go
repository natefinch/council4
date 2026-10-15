package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IntroductionToAnnihilation is the card definition for Introduction to Annihilation.
//
// Type: Sorcery — Lesson
// Cost: {5}
//
// Oracle text:
//
//	Exile target nonland permanent. Its controller draws a card.
var IntroductionToAnnihilation = newIntroductionToAnnihilation

func newIntroductionToAnnihilation() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Introduction to Annihilation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Lesson},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonland permanent",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target nonland permanent. Its controller draws a card.
		`,
		},
	}
}
