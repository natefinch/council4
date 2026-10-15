package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Perish is the card definition for Perish.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Destroy all green creatures. They can't be regenerated.
var Perish = newPerish

func newPerish() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Perish",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group:               game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green}}),
							PreventRegeneration: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all green creatures. They can't be regenerated.
		`,
		},
	}
}
