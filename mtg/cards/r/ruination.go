package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ruination is the card definition for Ruination.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Destroy all nonbasic lands.
var Ruination = newRuination

func newRuination() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ruination",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, ExcludedSupertype: types.Basic}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all nonbasic lands.
		`,
		},
	}
}
