package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MeasureOfWickedness is the card definition for Measure of Wickedness.
//
// Type: Enchantment
// Cost: {3}{B}
//
// Oracle text:
//
//	At the beginning of your end step, sacrifice this enchantment and you lose 8 life.
//	Whenever another card is put into your graveyard from anywhere, target opponent gains control of this enchantment.
var MeasureOfWickedness = newMeasureOfWickedness

func newMeasureOfWickedness() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Measure of Wickedness",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
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
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Sacrifice{
									Object: game.SourceCardPermanentReference(),
								},
							},
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(8),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:       game.EventZoneChanged,
							ExcludeSelf: true,
							Player:      game.TriggerPlayerYou,
							MatchToZone: true,
							ToZone:      zone.Graveyard,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourcePermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:            game.LayerControl,
											NewControllerRef: opt.Val(game.TargetPlayerReference(0)),
										},
									},
									Duration: game.DurationPermanent,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your end step, sacrifice this enchantment and you lose 8 life.
			Whenever another card is put into your graveyard from anywhere, target opponent gains control of this enchantment.
		`,
		},
	}
}
