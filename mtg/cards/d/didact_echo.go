package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DidactEcho is the card definition for Didact Echo.
//
// Type: Creature — Spirit Cleric
// Cost: {4}{U}
//
// Oracle text:
//
//	When this creature enters, draw a card.
//	Descend 4 — This creature has flying as long as there are four or more permanent cards in your graveyard.
var DidactEcho = newDidactEcho()

func newDidactEcho() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Didact Echo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit, types.Cleric},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardPermanentCardCount, Op: compare.GreaterOrEqual, Value: 4}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddKeywords: []game.Keyword{
								game.Flying,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, draw a card.
			Descend 4 — This creature has flying as long as there are four or more permanent cards in your graveyard.
		`,
		},
	}
}
