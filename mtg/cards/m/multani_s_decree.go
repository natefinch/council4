package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MultaniSDecree is the card definition for Multani's Decree.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Destroy all enchantments. You gain 2 life for each enchantment destroyed this way.
var MultaniSDecree = newMultaniSDecree

func newMultaniSDecree() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Multani's Decree",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),
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
			Destroy all enchantments. You gain 2 life for each enchantment destroyed this way.
		`,
		},
	}
}
