package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Hypochondria is the card definition for Hypochondria.
//
// Type: Enchantment
// Cost: {1}{W}
//
// Oracle text:
//
//	{W}, Discard a card: Prevent the next 3 damage that would be dealt to any target this turn.
//	{W}, Sacrifice this enchantment: Prevent the next 3 damage that would be dealt to any target this turn.
var Hypochondria = newHypochondria

func newHypochondria() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Hypochondria",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{W}, Discard a card: Prevent the next 3 damage that would be dealt to any target this turn.",
					ManaCost: opt.Val(cost.Mana{cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalDiscard,
							Text:   "Discard a card",
							Amount: 1,
							Source: zone.Hand,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									AnyTarget: game.AnyTargetDamageRecipient(0),
									Amount:    game.Fixed(3),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:     "{W}, Sacrifice this enchantment: Prevent the next 3 damage that would be dealt to any target this turn.",
					ManaCost: opt.Val(cost.Mana{cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this enchantment",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									AnyTarget: game.AnyTargetDamageRecipient(0),
									Amount:    game.Fixed(3),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{W}, Discard a card: Prevent the next 3 damage that would be dealt to any target this turn.
			{W}, Sacrifice this enchantment: Prevent the next 3 damage that would be dealt to any target this turn.
		`,
		},
	}
}
