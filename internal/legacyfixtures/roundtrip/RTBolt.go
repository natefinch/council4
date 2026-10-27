// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTBolt is the card definition for RT Bolt.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	RT Bolt deals 3 damage to any target.
var RTBolt = newRTBolt

func newRTBolt() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "RT Bolt",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			RT Bolt deals 3 damage to any target.
		`,
		},
	}
}
