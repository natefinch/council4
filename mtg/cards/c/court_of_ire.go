package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CourtOfIre is the card definition for Court of Ire.
//
// Type: Enchantment
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	When this enchantment enters, you become the monarch.
//	At the beginning of your upkeep, this enchantment deals 2 damage to any target. If you're the monarch, it deals 7 damage instead.
var CourtOfIre = newCourtOfIre

func newCourtOfIre() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Court of Ire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
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
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(2),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate:              true,
										ControllerIsMonarch: true,
									}),
								}),
							},
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(7),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControllerIsMonarch: true,
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this enchantment enters, you become the monarch.
			At the beginning of your upkeep, this enchantment deals 2 damage to any target. If you're the monarch, it deals 7 damage instead.
		`,
		},
	}
}
