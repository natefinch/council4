package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GhastlyDeathTyrant is the card definition for Ghastly Death Tyrant.
//
// Type: Creature — Beholder Skeleton
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Disintegration Ray — Destroy target enchantment an opponent controls. You lose life equal to its mana value.
//	• Death Ray — Creatures you control gain deathtouch until end of turn.
var GhastlyDeathTyrant = newGhastlyDeathTyrant()

func newGhastlyDeathTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ghastly Death Tyrant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beholder, types.Skeleton},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Disintegration Ray — Destroy target enchantment an opponent controls. You lose life equal to its mana value.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target enchantment an opponent controls",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}, Controller: game.ControllerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
									{
										Primitive: game.LoseLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountObjectManaValue,
												Multiplier: 1,
												Object:     game.TargetPermanentReference(0),
											}),
											Player: game.ControllerReference(),
										},
									},
								},
							},
							game.Mode{
								Text: "Death Ray — Creatures you control gain deathtouch until end of turn.",
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													AddKeywords: []game.Keyword{
														game.Deathtouch,
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			When this creature enters, choose one —
			• Disintegration Ray — Destroy target enchantment an opponent controls. You lose life equal to its mana value.
			• Death Ray — Creatures you control gain deathtouch until end of turn.
		`,
		},
	}
}
