package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShedWeakness is the card definition for Shed Weakness.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Target creature gets +2/+2 until end of turn. You may remove a -1/-1 counter from it.
var ShedWeakness = newShedWeakness

func newShedWeakness() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Shed Weakness",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
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
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(2),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.RemoveCounter{
							Amount:      game.Fixed(1),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.MinusOneMinusOne,
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets +2/+2 until end of turn. You may remove a -1/-1 counter from it.
		`,
		},
	}
}
