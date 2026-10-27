// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTDiscardCostSubject is the card definition for RT Discard Cost Subject.
//
// Type: Sorcery
//
// Oracle text:
//
//	As an additional cost to cast this spell, discard a card.
//	Draw a card. If the discarded card wasn't a land card, you gain 2 life.
var RTDiscardCostSubject = newRTDiscardCostSubject

func newRTDiscardCostSubject() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Discard Cost Subject",
			Types: []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:       cost.AdditionalDiscard,
					Text:       "discard a card",
					Amount:     1,
					Source:     zone.Hand,
					SubjectKey: "paid-cost-1-1",
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:        true,
								Object:        opt.Val(game.PaidCostReference("paid-cost-1-1", game.PaidCostDiscard, "")),
								ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, discard a card.
			Draw a card. If the discarded card wasn't a land card, you gain 2 life.
		`,
		},
	}
}
