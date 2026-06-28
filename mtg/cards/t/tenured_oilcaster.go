package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TenuredOilcaster is the card definition for Tenured Oilcaster.
//
// Type: Creature — Phyrexian Wizard
// Cost: {3}{B}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
//	This creature gets +3/+0 as long as an opponent has eight or more cards in their graveyard.
//	Whenever this creature attacks or blocks, each player mills a card.
var TenuredOilcaster = newTenuredOilcaster()

func newTenuredOilcaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Tenured Oilcaster",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateAnyOpponentGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 8}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     3,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventBlockerDeclared,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount:      game.Fixed(1),
									PlayerGroup: game.AllPlayersReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
			This creature gets +3/+0 as long as an opponent has eight or more cards in their graveyard.
			Whenever this creature attacks or blocks, each player mills a card.
		`,
		},
	}
}
