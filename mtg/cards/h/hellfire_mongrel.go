package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HellfireMongrel is the card definition for Hellfire Mongrel.
//
// Type: Creature — Elemental Dog
// Cost: {2}{R}
//
// Oracle text:
//
//	At the beginning of each opponent's upkeep, if that player has two or fewer cards in hand, this creature deals 2 damage to that player.
var HellfireMongrel = newHellfireMongrel

func newHellfireMongrel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hellfire Mongrel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Dog},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerOpponent,
							Step:       game.StepUpkeep,
						},
						InterveningIf: "if that player has two or fewer cards in hand",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateEventPlayerHandSize, Op: compare.LessOrEqual, Value: 2}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(2),
									Recipient:    game.PlayerDamageRecipient(game.EventPlayerReference()),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of each opponent's upkeep, if that player has two or fewer cards in hand, this creature deals 2 damage to that player.
		`,
		},
	}
}
