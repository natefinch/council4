package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ConsumingFervor is the card definition for Consuming Fervor.
//
// Type: Enchantment — Aura
// Cost: {R}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +3/+3 and has "At the beginning of your upkeep, put a -1/-1 counter on this creature."
var ConsumingFervor = newConsumingFervor

func newConsumingFervor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Consuming Fervor",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature gets +3/+3 and has \"At the beginning of your upkeep, put a -1/-1 counter on this creature.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     3,
							ToughnessDelta: 3,
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "At the beginning of your upkeep, put a -1/-1 counter on this creature.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerAt,
										Pattern: game.TriggerPattern{
											Event:           game.EventBeginningOfStep,
											Controller:      game.TriggerControllerYou,
											StackObjectKind: game.StackSpell,
											CounterKind:     counter.PlusOnePlusOne,
											Step:            game.StepUpkeep,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.AddCounter{
															Amount:      game.Fixed(1),
															Object:      game.SourcePermanentReference(),
															CounterKind: counter.MinusOneMinusOne,
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
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +3/+3 and has "At the beginning of your upkeep, put a -1/-1 counter on this creature."
		`,
		},
	}
}
