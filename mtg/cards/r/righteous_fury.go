package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RighteousFury is the card definition for Righteous Fury.
//
// Type: Sorcery
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Destroy all tapped creatures. You gain 2 life for each creature destroyed this way.
var RighteousFury = newRighteousFury

func newRighteousFury() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Righteous Fury",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriTrue}),
						},
						PublishResult: game.ResultKey("destroyed-this-way"),
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 2,
								ResultKey:  game.ResultKey("destroyed-this-way"),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all tapped creatures. You gain 2 life for each creature destroyed this way.
		`,
		},
	}
}
