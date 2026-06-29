package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AsylumVisitor is the card definition for Asylum Visitor.
//
// Type: Creature — Vampire Wizard
// Cost: {1}{B}
//
// Oracle text:
//
//	At the beginning of each player's upkeep, if that player has no cards in hand, you draw a card and you lose 1 life.
//	Madness {1}{B} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
var AsylumVisitor = newAsylumVisitor()

func newAsylumVisitor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Asylum Visitor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.MadnessKeyword{Cost: cost.Mana{cost.O(1), cost.B}},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepUpkeep,
						},
						InterveningIf: "if that player has no cards in hand",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateEventPlayerHandSize, Op: compare.LessOrEqual, Value: 0}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of each player's upkeep, if that player has no cards in hand, you draw a card and you lose 1 life.
			Madness {1}{B} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
		`,
		},
	}
}
