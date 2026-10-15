package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloodTithe is the card definition for Blood Tithe.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	Each opponent loses 3 life. You gain life equal to the life lost this way.
var BloodTithe = newBloodTithe

func newBloodTithe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Blood Tithe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.LoseLife{
							Amount:      game.Fixed(3),
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
			Each opponent loses 3 life. You gain life equal to the life lost this way.
		`,
		},
	}
}
