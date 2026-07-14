package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SyggRiverCutthroat is the card definition for Sygg, River Cutthroat.
//
// Type: Legendary Creature — Merfolk Rogue
// Cost: {U/B}{U/B}
//
// Oracle text:
//
//	At the beginning of each end step, if an opponent lost 3 or more life this turn, you may draw a card. (Damage causes loss of life.)
var SyggRiverCutthroat = newSyggRiverCutthroat

func newSyggRiverCutthroat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Sygg, River Cutthroat",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.U, mana.B),
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Merfolk, types.Rogue},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepEnd,
						},
						InterveningIf: "if an opponent lost 3 or more life this turn",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateAnyOpponentLifeLostThisTurn, Op: compare.GreaterOrEqual, Value: 3}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of each end step, if an opponent lost 3 or more life this turn, you may draw a card. (Damage causes loss of life.)
		`,
		},
	}
}
