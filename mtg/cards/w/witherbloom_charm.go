package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WitherbloomCharm is the card definition for Witherbloom Charm.
//
// Type: Instant
// Cost: {B}{G}
//
// Oracle text:
//
//	Choose one —
//	• You may sacrifice a permanent. If you do, draw two cards.
//	• You gain 5 life.
//	• Destroy target nonland permanent with mana value 2 or less.
var WitherbloomCharm = newWitherbloomCharm

func newWitherbloomCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Witherbloom Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "You may sacrifice a permanent. If you do, draw two cards.",
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Player: game.ControllerReference(),
									Amount: game.Fixed(1),
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
								LocalProducts: game.LocalProducts{
									Results: []game.ResultKey{game.ResultKey("if-you-do")},
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					},
					game.Mode{
						Text: "You gain 5 life.",
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(5),
									Player: game.ControllerReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy target nonland permanent with mana value 2 or less.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target nonland permanent with mana value 2 or less",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
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
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• You may sacrifice a permanent. If you do, draw two cards.
			• You gain 5 life.
			• Destroy target nonland permanent with mana value 2 or less.
		`,
		},
	}
}
