package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SheoldredSEdict is the card definition for Sheoldred's Edict.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	Choose one —
//	• Each opponent sacrifices a nontoken creature of their choice.
//	• Each opponent sacrifices a creature token of their choice.
//	• Each opponent sacrifices a planeswalker of their choice.
var SheoldredSEdict = newSheoldredSEdict

func newSheoldredSEdict() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sheoldred's Edict",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Each opponent sacrifices a nontoken creature of their choice.",
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsReference(),
									Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
								},
							},
						},
					},
					game.Mode{
						Text: "Each opponent sacrifices a creature token of their choice.",
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsReference(),
									Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true},
								},
							},
						},
					},
					game.Mode{
						Text: "Each opponent sacrifices a planeswalker of their choice.",
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsReference(),
									Selection:   game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
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
			• Each opponent sacrifices a nontoken creature of their choice.
			• Each opponent sacrifices a creature token of their choice.
			• Each opponent sacrifices a planeswalker of their choice.
		`,
		},
	}
}
