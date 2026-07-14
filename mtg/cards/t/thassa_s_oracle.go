package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThassaSOracle is the card definition for Thassa's Oracle.
//
// Type: Creature — Merfolk Wizard
// Cost: {U}{U}
//
// Oracle text:
//
//	When this creature enters, look at the top X cards of your library, where X is your devotion to blue. Put up to one of them on top of your library and the rest on the bottom of your library in a random order. If X is greater than or equal to the number of cards in your library, you win the game. (Each {U} in the mana costs of permanents you control counts toward your devotion to blue.)
var ThassaSOracle = newThassaSOracle

func newThassaSOracle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Thassa's Oracle",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
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
						Text: "When this creature enters, look at the top X cards of your library, where X is your devotion to blue. Put up to one of them on top of your library and the rest on the bottom of your library in a random order. If X is greater than or equal to the number of cards in your library, you win the game. (Each {U} in the mana costs of permanents you control counts toward your devotion to blue.)",
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player: game.ControllerReference(),
									Look: game.Dynamic(game.DynamicAmount{
										Kind:   game.DynamicAmountDevotion,
										Colors: []color.Color{color.Blue},
									}),
									Take:        game.Fixed(1),
									Remainder:   game.DigRemainderLibraryBottom,
									TakeUpTo:    true,
									Destination: zone.Library,
								},
							},
							{
								Primitive: game.PlayerWinsGame{
									Player: game.ControllerReference(),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLibrarySize, Op: compare.LessOrEqual, Value: 0, ValueAmount: opt.Val(game.DynamicAmount{
											Kind:   game.DynamicAmountDevotion,
											Colors: []color.Color{color.Blue},
										})}},
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, look at the top X cards of your library, where X is your devotion to blue. Put up to one of them on top of your library and the rest on the bottom of your library in a random order. If X is greater than or equal to the number of cards in your library, you win the game. (Each {U} in the mana costs of permanents you control counts toward your devotion to blue.)
		`,
		},
	}
}
