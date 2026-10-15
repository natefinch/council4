package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Unnerve is the card definition for Unnerve.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	Each opponent discards two cards.
var Unnerve = newUnnerve

func newUnnerve() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Unnerve",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount:      game.Fixed(2),
							PlayerGroup: game.OpponentsReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each opponent discards two cards.
		`,
		},
	}
}
