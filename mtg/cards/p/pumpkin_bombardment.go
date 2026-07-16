package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PumpkinBombardment is the card definition for Pumpkin Bombardment.
//
// Type: Sorcery
// Cost: {B/R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, discard a card or pay {2}.
//	Pumpkin Bombardment deals 3 damage to target creature.
var PumpkinBombardment = newPumpkinBombardment

func newPumpkinBombardment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Pumpkin Bombardment",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.R),
			}),
			Colors: []color.Color{color.Black, color.Red},
			Types:  []types.Card{types.Sorcery},
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
							Label: "Pay {2}",
							Mana:  cost.Mana{cost.O(2)},
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
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, discard a card or pay {2}.
			Pumpkin Bombardment deals 3 damage to target creature.
		`,
		},
	}
}
