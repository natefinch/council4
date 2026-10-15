package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GrabThePrize is the card definition for Grab the Prize.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, discard a card.
//	Draw two cards. If the discarded card wasn't a land card, Grab the Prize deals 2 damage to each opponent.
var GrabThePrize = newGrabThePrize

func newGrabThePrize() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Grab the Prize",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
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
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
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
			Draw two cards. If the discarded card wasn't a land card, Grab the Prize deals 2 damage to each opponent.
		`,
		},
	}
}
