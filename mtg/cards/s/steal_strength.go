package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StealStrength is the card definition for Steal Strength.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	Target creature gets +1/+1 until end of turn. Another target creature gets -1/-1 until end of turn.
var StealStrength = newStealStrength

func newStealStrength() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Steal Strength",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
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
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(1),
							ToughnessDelta: game.Fixed(1),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(1),
							PowerDelta:     game.Fixed(-1),
							ToughnessDelta: game.Fixed(-1),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets +1/+1 until end of turn. Another target creature gets -1/-1 until end of turn.
		`,
		},
	}
}
