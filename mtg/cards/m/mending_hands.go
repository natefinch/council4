package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MendingHands is the card definition for Mending Hands.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Prevent the next 4 damage that would be dealt to any target this turn.
var MendingHands = newMendingHands

func newMendingHands() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Mending Hands",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PreventDamage{
							AnyTarget: game.AnyTargetDamageRecipient(0),
							Amount:    game.Fixed(4),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Prevent the next 4 damage that would be dealt to any target this turn.
		`,
		},
	}
}
