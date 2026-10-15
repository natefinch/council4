package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DemandAnswers is the card definition for Demand Answers.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice an artifact or discard a card.
//	Draw two cards.
var DemandAnswers = newDemandAnswers

func newDemandAnswers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Demand Answers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice an artifact",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Artifact,
					ChoiceGroup:        1,
				},
				{
					Kind:        cost.AdditionalDiscard,
					Text:        "discard a card",
					Amount:      1,
					Source:      zone.Hand,
					ChoiceGroup: 1,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice an artifact or discard a card.
			Draw two cards.
		`,
		},
	}
}
