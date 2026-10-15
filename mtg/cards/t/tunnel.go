package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Tunnel is the card definition for Tunnel.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Destroy target Wall. It can't be regenerated.
var Tunnel = newTunnel

func newTunnel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Tunnel",
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
						Constraint: "target Wall",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Wall")}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object:              game.TargetPermanentReference(0),
							PreventRegeneration: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target Wall. It can't be regenerated.
		`,
		},
	}
}
