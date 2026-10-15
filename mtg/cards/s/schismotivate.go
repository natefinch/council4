package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Schismotivate is the card definition for Schismotivate.
//
// Type: Instant
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	Target creature gets +4/+0 until end of turn. Another target creature gets -4/-0 until end of turn.
var Schismotivate = newSchismotivate

func newSchismotivate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Schismotivate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Colors: []color.Color{color.Red, color.Blue},
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
					game.TargetSpec{
						MinTargets:               1,
						MaxTargets:               1,
						Constraint:               "another target creature",
						Allow:                    game.TargetAllowPermanent,
						Selection:                opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
						DistinctFromPriorTargets: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:     game.TargetPermanentReference(0),
							PowerDelta: game.Fixed(4),
							Duration:   game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ModifyPT{
							Object:     game.TargetPermanentReference(1),
							PowerDelta: game.Fixed(-4),
							Duration:   game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets +4/+0 until end of turn. Another target creature gets -4/-0 until end of turn.
		`,
		},
	}
}
