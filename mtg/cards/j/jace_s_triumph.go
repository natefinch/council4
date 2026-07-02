package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JaceSTriumph is the card definition for Jace's Triumph.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	Draw two cards. If you control a Jace planeswalker, draw three cards instead.
var JaceSTriumph = newJaceSTriumph()

func newJaceSTriumph() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Jace's Triumph",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate: true,
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Sub("Jace")}},
								}),
							}),
						}),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Sub("Jace")}},
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Draw two cards. If you control a Jace planeswalker, draw three cards instead.
		`,
		},
	}
}
