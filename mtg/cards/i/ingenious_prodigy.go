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

// IngeniousProdigy is the card definition for Ingenious Prodigy.
//
// Type: Creature — Human Wizard
// Cost: {X}{U}
//
// Oracle text:
//
//	Skulk (This creature can't be blocked by creatures with greater power.)
//	This creature enters with X +1/+1 counters on it.
//	At the beginning of your upkeep, if this creature has one or more +1/+1 counters on it, you may remove a +1/+1 counter from it. If you do, draw a card.
var IngeniousProdigy = newIngeniousProdigy()

func newIngeniousProdigy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ingenious Prodigy",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.SkulkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
						InterveningIf: "if this creature has one or more +1/+1 counters on it",
						InterveningCondition: opt.Val(game.Condition{
							Object:        opt.Val(game.SourcePermanentReference()),
							ObjectMatches: opt.Val(game.Selection{RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 1}), RequiredCounter: counter.PlusOnePlusOne}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RemoveCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
			},
			OracleText: `
			Skulk (This creature can't be blocked by creatures with greater power.)
			This creature enters with X +1/+1 counters on it.
			At the beginning of your upkeep, if this creature has one or more +1/+1 counters on it, you may remove a +1/+1 counter from it. If you do, draw a card.
		`,
		},
	}
}
