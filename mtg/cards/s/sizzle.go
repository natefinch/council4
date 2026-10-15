package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Sizzle is the card definition for Sizzle.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Sizzle deals 3 damage to each opponent.
var Sizzle = newSizzle

func newSizzle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Sizzle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Sizzle deals 3 damage to each opponent.
		`,
		},
	}
}
