package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KolaghanSCommand is the card definition for Kolaghan's Command.
//
// Type: Instant
// Cost: {1}{B}{R}
//
// Oracle text:
//
//	Choose two —
//	• Return target creature card from your graveyard to your hand.
//	• Target player discards a card.
//	• Destroy target artifact.
//	• Kolaghan's Command deals 2 damage to any target.
var KolaghanSCommand = newKolaghanSCommand

func newKolaghanSCommand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Kolaghan's Command",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.R,
			}),
			Colors: []color.Color{color.Black, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Return target creature card from your graveyard to your hand.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card: game.CardReference{
										Kind: game.CardReferenceTarget,
									},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
							},
						},
					},
					game.Mode{
						Text: "Target player discards a card.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy target artifact.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Kolaghan's Command deals 2 damage to any target.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(2),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
				},
				MinModes: 2,
				MaxModes: 2,
			}),
			OracleText: `
			Choose two —
			• Return target creature card from your graveyard to your hand.
			• Target player discards a card.
			• Destroy target artifact.
			• Kolaghan's Command deals 2 damage to any target.
		`,
		},
	}
}
