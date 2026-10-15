package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Twiddle is the card definition for Twiddle.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	You may tap or untap target artifact, creature, or land.
var Twiddle = newTwiddle

func newTwiddle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Twiddle",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact, creature, or land",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.TapOrUntap{
							Object: game.TargetPermanentReference(0),
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			You may tap or untap target artifact, creature, or land.
		`,
		},
	}
}
