package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Depressurize is the card definition for Depressurize.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	Target creature gets -3/-0 until end of turn. Then if that creature's power is 0 or less, destroy it.
var Depressurize = newDepressurize

func newDepressurize() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Depressurize",
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
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:     game.TargetPermanentReference(0),
							PowerDelta: game.Fixed(-3),
							Duration:   game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.TargetPermanentReference(0)),
								ObjectMatches: opt.Val(game.Selection{Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 0})}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets -3/-0 until end of turn. Then if that creature's power is 0 or less, destroy it.
		`,
		},
	}
}
