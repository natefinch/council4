package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cleanfall is the card definition for Cleanfall.
//
// Type: Sorcery — Arcane
// Cost: {2}{W}
//
// Oracle text:
//
//	Destroy all enchantments.
var Cleanfall = newCleanfall

func newCleanfall() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Cleanfall",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Arcane},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all enchantments.
		`,
		},
	}
}
