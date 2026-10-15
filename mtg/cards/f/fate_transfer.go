package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FateTransfer is the card definition for Fate Transfer.
//
// Type: Instant
// Cost: {1}{U/B}
//
// Oracle text:
//
//	Move all counters from target creature onto another target creature.
var FateTransfer = newFateTransfer

func newFateTransfer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Fate Transfer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors: []color.Color{color.Black, color.Blue},
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
						Primitive: game.MoveCounters{
							Object:      game.TargetPermanentReference(1),
							CounterKind: counter.PlusOnePlusOne,
							Source: game.CounterSourceSpec{
								Kind:   game.CounterSourceTarget,
								Object: game.TargetPermanentReference(0),
							},
							AllKinds: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Move all counters from target creature onto another target creature.
		`,
		},
	}
}
