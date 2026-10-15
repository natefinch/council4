package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MaestrosCharm is the card definition for Maestros Charm.
//
// Type: Instant
// Cost: {U}{B}{R}
//
// Oracle text:
//
//	Choose one —
//	• Look at the top five cards of your library. Put one of those cards into your hand and the rest into your graveyard.
//	• Each opponent loses 3 life and you gain 3 life.
//	• Maestros Charm deals 5 damage to target creature or planeswalker.
var MaestrosCharm = newMaestrosCharm

func newMaestrosCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Maestros Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
				cost.R,
			}),
			Colors: []color.Color{color.Black, color.Red, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Look at the top five cards of your library. Put one of those cards into your hand and the rest into your graveyard.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player: game.ControllerReference(),
									Look:   game.Fixed(5),
									Take:   game.Fixed(1),
								},
							},
						},
					},
					game.Mode{
						Text: "Each opponent loses 3 life and you gain 3 life.",
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount:      game.Fixed(3),
									PlayerGroup: game.OpponentsReference(),
								},
								PublishResult: game.ResultKey("life-change"),
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Maestros Charm deals 5 damage to target creature or planeswalker.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature or planeswalker",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
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
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Look at the top five cards of your library. Put one of those cards into your hand and the rest into your graveyard.
			• Each opponent loses 3 life and you gain 3 life.
			• Maestros Charm deals 5 damage to target creature or planeswalker.
		`,
		},
	}
}
