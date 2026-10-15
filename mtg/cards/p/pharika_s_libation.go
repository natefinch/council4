package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PharikaSLibation is the card definition for Pharika's Libation.
//
// Type: Instant
// Cost: {2}{B}
//
// Oracle text:
//
//	Choose one —
//	• Target opponent sacrifices a creature of their choice.
//	• Target opponent sacrifices an enchantment of their choice.
var PharikaSLibation = newPharikaSLibation

func newPharikaSLibation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Pharika's Libation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target opponent sacrifices a creature of their choice.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Player:    game.TargetPlayerReference(0),
									Amount:    game.Fixed(1),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
							},
						},
					},
					game.Mode{
						Text: "Target opponent sacrifices an enchantment of their choice.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Player:    game.TargetPlayerReference(0),
									Amount:    game.Fixed(1),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
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
			• Target opponent sacrifices a creature of their choice.
			• Target opponent sacrifices an enchantment of their choice.
		`,
		},
	}
}
