package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NicolBolasTheDeceiver is the card definition for Nicol Bolas, the Deceiver.
//
// Type: Legendary Planeswalker — Bolas
// Cost: {5}{U}{B}{R}
//
// Oracle text:
//
//	+3: Each opponent loses 3 life unless that player sacrifices a nonland permanent of their choice or discards a card.
//	−3: Destroy target creature. Draw a card.
//	−11: Nicol Bolas deals 7 damage to each opponent. You draw seven cards.
var NicolBolasTheDeceiver = newNicolBolasTheDeceiver

func newNicolBolasTheDeceiver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Nicol Bolas, the Deceiver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Bolas},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+3: Each opponent loses 3 life unless that player sacrifices a nonland permanent of their choice or discards a card.",
					LoyaltyCost: 3,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PunisherEachLoseLife{
											PlayerGroup:        game.OpponentsReference(),
											Amount:             game.Fixed(3),
											AllowSacrifice:     true,
											SacrificeSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
											AllowDiscard:       true,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−3: Destroy target creature. Draw a card.",
					LoyaltyCost: -3,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
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
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−11: Nicol Bolas deals 7 damage to each opponent. You draw seven cards.",
					LoyaltyCost: -11,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(7),
											Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(7),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			+3: Each opponent loses 3 life unless that player sacrifices a nonland permanent of their choice or discards a card.
			−3: Destroy target creature. Draw a card.
			−11: Nicol Bolas deals 7 damage to each opponent. You draw seven cards.
		`,
		},
	}
}
