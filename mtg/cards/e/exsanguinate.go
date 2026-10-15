package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Exsanguinate is the card definition for Exsanguinate.
//
// Type: Sorcery
// Cost: {X}{B}{B}
//
// Oracle text:
//
//	Each opponent loses X life. You gain life equal to the life lost this way.
var Exsanguinate = newExsanguinate

func newExsanguinate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Exsanguinate",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.LoseLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							PlayerGroup: game.OpponentsReference(),
						},
						PublishResult: game.ResultKey("life-change"),
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:      game.DynamicAmountPreviousEffectResult,
								ResultKey: game.ResultKey("life-change"),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each opponent loses X life. You gain life equal to the life lost this way.
		`,
		},
	}
}
