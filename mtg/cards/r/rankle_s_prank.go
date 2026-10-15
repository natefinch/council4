package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RankleSPrank is the card definition for Rankle's Prank.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Choose one or more —
//	• Each player discards two cards.
//	• Each player loses 4 life.
//	• Each player sacrifices two creatures of their choice.
var RankleSPrank = newRankleSPrank

func newRankleSPrank() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Rankle's Prank",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Each player discards two cards.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount:      game.Fixed(2),
									PlayerGroup: game.AllPlayersReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Each player loses 4 life.",
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount:      game.Fixed(4),
									PlayerGroup: game.AllPlayersReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Each player sacrifices two creatures of their choice.",
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									PlayerGroup: game.AllPlayersReference(),
									Amount:      game.Fixed(2),
									Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 3,
			}),
			OracleText: `
			Choose one or more —
			• Each player discards two cards.
			• Each player loses 4 life.
			• Each player sacrifices two creatures of their choice.
		`,
		},
	}
}
