package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Anarchy is the card definition for Anarchy.
//
// Type: Sorcery
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Destroy all white permanents.
var Anarchy = newAnarchy

func newAnarchy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Anarchy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{ColorsAny: []color.Color{color.White}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all white permanents.
		`,
		},
	}
}
