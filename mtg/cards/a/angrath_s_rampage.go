package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AngrathSRampage is the card definition for Angrath's Rampage.
//
// Type: Sorcery
// Cost: {B}{R}
//
// Oracle text:
//
//	Choose one —
//	• Target player sacrifices an artifact of their choice.
//	• Target player sacrifices a creature of their choice.
//	• Target player sacrifices a planeswalker of their choice.
var AngrathSRampage = newAngrathSRampage

func newAngrathSRampage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Angrath's Rampage",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.R,
			}),
			Colors: []color.Color{color.Black, color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target player sacrifices an artifact of their choice.",
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
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.TargetPlayerReference(0),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
								},
							},
						},
					},
					game.Mode{
						Text: "Target player sacrifices a creature of their choice.",
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
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.TargetPlayerReference(0),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
							},
						},
					},
					game.Mode{
						Text: "Target player sacrifices a planeswalker of their choice.",
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
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.TargetPlayerReference(0),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
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
			• Target player sacrifices an artifact of their choice.
			• Target player sacrifices a creature of their choice.
			• Target player sacrifices a planeswalker of their choice.
		`,
		},
	}
}
