package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EbonyOwlNetsuke is the card definition for Ebony Owl Netsuke.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	At the beginning of each opponent's upkeep, if that player has seven or more cards in hand, this artifact deals 4 damage to that player.
var EbonyOwlNetsuke = newEbonyOwlNetsuke()

func newEbonyOwlNetsuke() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ebony Owl Netsuke",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerOpponent,
							Step:       game.StepUpkeep,
						},
						InterveningIf: "if that player has seven or more cards in hand",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateEventPlayerHandSize, Op: compare.GreaterOrEqual, Value: 7}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(4),
									Recipient:    game.PlayerDamageRecipient(game.EventPlayerReference()),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of each opponent's upkeep, if that player has seven or more cards in hand, this artifact deals 4 damage to that player.
		`,
		},
	}
}
