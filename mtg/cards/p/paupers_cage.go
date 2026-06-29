package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PaupersCage is the card definition for Paupers' Cage.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	At the beginning of each opponent's upkeep, if that player has two or fewer cards in hand, this artifact deals 2 damage to that player.
var PaupersCage = newPaupersCage()

func newPaupersCage() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Paupers' Cage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
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
			At the beginning of each opponent's upkeep, if that player has two or fewer cards in hand, this artifact deals 2 damage to that player.
		`,
		},
	}
}
