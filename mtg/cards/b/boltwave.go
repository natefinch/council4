package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Boltwave is the card definition for Boltwave.
//
// Type: Sorcery
// Cost: {R}
//
// Oracle text:
//
//	Boltwave deals 3 damage to each opponent.
var Boltwave = newBoltwave

func newBoltwave() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Boltwave",
			ManaCost: opt.Val(cost.Mana{
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
			Boltwave deals 3 damage to each opponent.
		`,
		},
	}
}
