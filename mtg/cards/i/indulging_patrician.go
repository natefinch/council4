package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IndulgingPatrician is the card definition for Indulging Patrician.
//
// Type: Creature — Vampire Noble
// Cost: {1}{W}{B}
//
// Oracle text:
//
//	Flying
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
//	At the beginning of your end step, if you gained 3 or more life this turn, each opponent loses 3 life.
var IndulgingPatrician = newIndulgingPatrician

func newIndulgingPatrician() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Indulging Patrician",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Noble},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your end step, if you gained 3 or more life this turn, each opponent loses 3 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepEnd,
						},
						InterveningIf: "if you gained 3 or more life this turn",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if you gained 3 or more life this turn",
							Aggregates: []game.AggregateComparison{
								game.AggregateComparison{
									Aggregate: game.AggregateControllerGainedLifeThisTurn,
									Op:        compare.GreaterOrEqual,
									Value:     3,
								},
							},
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(3),
											PlayerGroup: game.OpponentsReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Flying
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
			At the beginning of your end step, if you gained 3 or more life this turn, each opponent loses 3 life.
		`,
		},
	}
}
