package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnexplainedVision is the card definition for Unexplained Vision.
//
// Type: Sorcery
// Cost: {4}{U}
//
// Oracle text:
//
//	Draw three cards.
//	Adamant — If at least three blue mana was spent to cast this spell, scry 3.
var UnexplainedVision = newUnexplainedVision

func newUnexplainedVision() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Unexplained Vision",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.Blue, Count: 3},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Draw three cards.
			Adamant — If at least three blue mana was spent to cast this spell, scry 3.
		`,
		},
	}
}
