package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrismariCharm is the card definition for Prismari Charm.
//
// Type: Instant
// Cost: {U}{R}
//
// Oracle text:
//
//	Choose one —
//	• Surveil 2, then draw a card.
//	• Prismari Charm deals 1 damage to each of one or two targets.
//	• Return target nonland permanent to its owner's hand.
var PrismariCharm = newPrismariCharm

func newPrismariCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Prismari Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
			}),
			Colors: []color.Color{color.Red, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Surveil 2, then draw a card.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Surveil{
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
					},
					game.Mode{
						Text: "Prismari Charm deals 1 damage to each of one or two targets.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 2,
								Constraint: "one or two targets",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(1),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(1),
									Recipient: game.AnyTargetDamageRecipient(1),
								},
							},
						},
					},
					game.Mode{
						Text: "Return target nonland permanent to its owner's hand.",
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
								Primitive: game.Bounce{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Surveil 2, then draw a card.
			• Prismari Charm deals 1 damage to each of one or two targets.
			• Return target nonland permanent to its owner's hand.
		`,
		},
	}
}
