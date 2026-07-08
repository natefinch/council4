package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShriekingAffliction is the card definition for Shrieking Affliction.
//
// Type: Enchantment
// Cost: {B}
//
// Oracle text:
//
//	At the beginning of each opponent's upkeep, if that player has one or fewer cards in hand, they lose 3 life.
var ShriekingAffliction = newShriekingAffliction

func newShriekingAffliction() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Shrieking Affliction",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerOpponent,
							Step:       game.StepUpkeep,
						},
						InterveningIf: "if that player has one or fewer cards in hand",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateEventPlayerHandSize, Op: compare.LessOrEqual, Value: 1}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(3),
									Player: game.EventPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of each opponent's upkeep, if that player has one or fewer cards in hand, they lose 3 life.
		`,
		},
	}
}
