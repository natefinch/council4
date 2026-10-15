package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SinkingFeeling is the card definition for Sinking Feeling.
//
// Type: Enchantment — Aura
// Cost: {2}{U}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature doesn't untap during its controller's untap step.
//	Enchanted creature has "{1}, Put a -1/-1 counter on this creature: Untap this creature."
var SinkingFeeling = newSinkingFeeling

func newSinkingFeeling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sinking Feeling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
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
					Text: "Enchanted creature doesn't untap during its controller's untap step.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectDoesntUntap,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
						},
					},
				},
				game.StaticAbility{
					Text: "Enchanted creature has \"{1}, Put a -1/-1 counter on this creature: Untap this creature.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{1}, Put a -1/-1 counter on this creature: Untap this creature.",
									ManaCost: opt.Val(cost.Mana{
										cost.O(1),
									}),
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:        cost.AdditionalPutCounter,
											Text:        "Put a -1/-1 counter on this creature",
											Amount:      1,
											CounterKind: counter.MinusOneMinusOne,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.Untap{
															Object: game.SourcePermanentReference(),
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
			Enchanted creature doesn't untap during its controller's untap step.
			Enchanted creature has "{1}, Put a -1/-1 counter on this creature: Untap this creature."
		`,
		},
	}
}
