package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Tremble is the card definition for Tremble.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Each player sacrifices a land of their choice.
var Tremble = newTremble

func newTremble() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Tremble",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							PlayerGroup: game.AllPlayersReference(),
							Amount:      game.Fixed(1),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player sacrifices a land of their choice.
		`,
		},
	}
}
