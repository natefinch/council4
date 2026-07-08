package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Recoil is the card definition for Recoil.
//
// Type: Instant
// Cost: {1}{U}{B}
//
// Oracle text:
//
//	Return target permanent to its owner's hand. Then that player discards a card.
var Recoil = newRecoil

func newRecoil() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Recoil",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target permanent",
						Allow:      game.TargetAllowPermanent,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Bounce{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target permanent to its owner's hand. Then that player discards a card.
		`,
		},
	}
}
