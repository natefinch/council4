package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AbzanCharm is the card definition for Abzan Charm.
//
// Type: Instant
// Cost: {W}{B}{G}
//
// Oracle text:
//
//	Choose one —
//	• Exile target creature with power 3 or greater.
//	• You draw two cards and you lose 2 life.
//	• Distribute two +1/+1 counters among one or two target creatures.
var AbzanCharm = newAbzanCharm

func newAbzanCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Abzan Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Exile target creature with power 3 or greater.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature with power 3 or greater",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "You draw two cards and you lose 2 life.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Distribute two +1/+1 counters among one or two target creatures.",
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
									CounterKind: counter.PlusOnePlusOne,
									Distribute:  true,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Exile target creature with power 3 or greater.
			• You draw two cards and you lose 2 life.
			• Distribute two +1/+1 counters among one or two target creatures.
		`,
		},
	}
}
