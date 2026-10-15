package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BreathOfMalfegor is the card definition for Breath of Malfegor.
//
// Type: Instant
// Cost: {3}{B}{R}
//
// Oracle text:
//
//	Breath of Malfegor deals 5 damage to each opponent.
var BreathOfMalfegor = newBreathOfMalfegor

func newBreathOfMalfegor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Breath of Malfegor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.R,
			}),
			Colors: []color.Color{color.Black, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Breath of Malfegor deals 5 damage to each opponent.
		`,
		},
	}
}
