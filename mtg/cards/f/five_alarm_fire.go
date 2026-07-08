package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FiveAlarmFire is the card definition for Five-Alarm Fire.
//
// Type: Enchantment
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Whenever a creature you control deals combat damage, put a blaze counter on this enchantment.
//	Remove five blaze counters from this enchantment: It deals 5 damage to any target.
var FiveAlarmFire = newFiveAlarmFire

func newFiveAlarmFire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Five-Alarm Fire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove five blaze counters from this enchantment: It deals 5 damage to any target.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove five blaze counters from this enchantment",
							Amount:      5,
							CounterKind: counter.Blaze,
						},
					},
					ZoneOfFunction: zone.Battlefield,
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
									Amount:       game.Fixed(5),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							RequireCombatDamage:   true,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Blaze,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature you control deals combat damage, put a blaze counter on this enchantment.
			Remove five blaze counters from this enchantment: It deals 5 damage to any target.
		`,
		},
	}
}
