package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IndestructibleAura is the card definition for Indestructible Aura.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Prevent all damage that would be dealt to target creature this turn.
var IndestructibleAura = newIndestructibleAura

func newIndestructibleAura() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Indestructible Aura",
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
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PreventDamage{
							Object: game.TargetPermanentReference(0),
							All:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Prevent all damage that would be dealt to target creature this turn.
		`,
		},
	}
}
