package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Jace is the card definition for Jace.
//
// Type: Token Planeswalker — Jace
//
// Oracle text:
//   −1: Surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)
//   −3: Draw a card.

// JaceTokena6a09b69dd864fda8971b77a2a4abddb is the card definition for Jace.
var JaceTokena6a09b69dd864fda8971b77a2a4abddb = newJaceTokena6a09b69dd864fda8971b77a2a4abddb()

func newJaceTokena6a09b69dd864fda8971b77a2a4abddb() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:     "Jace",
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Planeswalker},
			Subtypes: []types.Sub{types.Jace},
			Loyalty:  opt.Val(0),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "−1: Surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)",
					LoyaltyCost: -1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Surveil{
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
					Text:        "−3: Draw a card.",
					LoyaltyCost: -3,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
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
			},
			OracleText: `
			−1: Surveil 1. (Look at the top card of your library. You may put that card into your graveyard.)
			−3: Draw a card.
		`,
		},
	}
}
