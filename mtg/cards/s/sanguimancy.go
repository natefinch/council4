package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Sanguimancy is the card definition for Sanguimancy.
//
// Type: Sorcery
// Cost: {4}{B}
//
// Oracle text:
//
//	You draw X cards and you lose X life, where X is your devotion to black. (Each {B} in the mana costs of permanents you control counts toward your devotion to black.)
var Sanguimancy = newSanguimancy

func newSanguimancy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sanguimancy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountDevotion,
								Multiplier: 1,
								Colors:     []color.Color{color.Black},
							}),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountDevotion,
								Multiplier: 1,
								Colors:     []color.Color{color.Black},
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You draw X cards and you lose X life, where X is your devotion to black. (Each {B} in the mana costs of permanents you control counts toward your devotion to black.)
		`,
		},
	}
}
