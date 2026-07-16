package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LightningAxe is the card definition for Lightning Axe.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, discard a card or pay {5}.
//	Lightning Axe deals 5 damage to target creature.
var LightningAxe = newLightningAxe

func newLightningAxe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Lightning Axe",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			AdditionalCostChoices: []cost.AdditionalChoice{
				cost.AdditionalChoice{
					Options: []cost.AdditionalChoiceOption{
						cost.AdditionalChoiceOption{
							Label: "Discard a card",
							Costs: []cost.Additional{
								{
									Kind:   cost.AdditionalDiscard,
									Text:   "discard a card",
									Amount: 1,
									Source: zone.Hand,
								},
							},
						},
						cost.AdditionalChoiceOption{
							Label: "Pay {5}",
							Mana:  cost.Mana{cost.O(5)},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, discard a card or pay {5}.
			Lightning Axe deals 5 damage to target creature.
		`,
		},
	}
}
