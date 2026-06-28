package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DemolisherSpawn is the card definition for Demolisher Spawn.
//
// Type: Enchantment Creature — Horror
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Trample, haste
//	Delirium — Whenever this creature attacks, if there are four or more card types among cards in your graveyard, other attacking creatures get +4/+4 until end of turn.
var DemolisherSpawn = newDemolisherSpawn()

func newDemolisherSpawn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Demolisher Spawn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Horror},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if there are four or more card types among cards in your graveyard",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardCardTypeCount, Op: compare.GreaterOrEqual, Value: 4}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}, game.SourcePermanentReference()),
											PowerDelta:     4,
											ToughnessDelta: 4,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample, haste
			Delirium — Whenever this creature attacks, if there are four or more card types among cards in your graveyard, other attacking creatures get +4/+4 until end of turn.
		`,
		},
	}
}
