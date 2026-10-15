package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FelidarSovereign is the card definition for Felidar Sovereign.
//
// Type: Creature — Cat Beast
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Vigilance (Attacking doesn't cause this creature to tap.)
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
//	At the beginning of your upkeep, if you have 40 or more life, you win the game.
var FelidarSovereign = newFelidarSovereign

func newFelidarSovereign() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Felidar Sovereign",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.LifelinkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, if you have 40 or more life, you win the game.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
						InterveningIf: "if you have 40 or more life",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if you have 40 or more life",
							Aggregates: []game.AggregateComparison{
								game.AggregateComparison{
									Aggregate: game.AggregateControllerLife,
									Op:        compare.GreaterOrEqual,
									Value:     40,
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
										Primitive: game.PlayerWinsGame{
											Player: game.ControllerReference(),
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
			Vigilance (Attacking doesn't cause this creature to tap.)
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
			At the beginning of your upkeep, if you have 40 or more life, you win the game.
		`,
		},
	}
}
