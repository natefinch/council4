package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RakdosCharm is the card definition for Rakdos Charm.
//
// Type: Instant
// Cost: {B}{R}
//
// Oracle text:
//
//	Choose one —
//	• Exile target player's graveyard.
//	• Destroy target artifact.
//	• Each creature deals 1 damage to its controller.
var RakdosCharm = newRakdosCharm

func newRakdosCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Rakdos Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.R,
			}),
			Colors: []color.Color{color.Black, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Exile target player's graveyard.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player's graveyard",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Player:      game.TargetPlayerReference(0),
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
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
						Text: "Each creature deals 1 damage to its controller.",
						Sequence: []game.Instruction{
							{
								Primitive: game.GroupSourceDamage{
									Group:  game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									Amount: game.Fixed(1),
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
			• Exile target player's graveyard.
			• Destroy target artifact.
			• Each creature deals 1 damage to its controller.
		`,
		},
	}
}
