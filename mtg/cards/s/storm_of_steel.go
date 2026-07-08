package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StormOfSteel is the card definition for Storm of Steel.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Storm of Steel deals 2 damage to each of one or two targets.
var StormOfSteel = newStormOfSteel

func newStormOfSteel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Storm of Steel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 2,
						Constraint: "one or two targets",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.AnyTargetDamageRecipient(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Storm of Steel deals 2 damage to each of one or two targets.
		`,
		},
	}
}
