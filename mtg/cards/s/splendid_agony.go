package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SplendidAgony is the card definition for Splendid Agony.
//
// Type: Instant
// Cost: {2}{B}
//
// Oracle text:
//
//	Distribute two -1/-1 counters among one or two target creatures.
var SplendidAgony = newSplendidAgony

func newSplendidAgony() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Splendid Agony",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 2,
						Constraint: "one or two target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(2),
							Object:      game.AllTargetPermanentsReference(0),
							CounterKind: counter.MinusOneMinusOne,
							Distribute:  true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Distribute two -1/-1 counters among one or two target creatures.
		`,
		},
	}
}
