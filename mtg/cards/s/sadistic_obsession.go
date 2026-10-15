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

// SadisticObsession is the card definition for Sadistic Obsession.
//
// Type: Enchantment — Aura
// Cost: {3}{B}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has "{B}, {T}: Put a -1/-1 counter on target creature."
var SadisticObsession = newSadisticObsession

func newSadisticObsession() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sadistic Obsession",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
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
					Text: "Enchanted creature has \"{B}, {T}: Put a -1/-1 counter on target creature.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{B}, {T}: Put a -1/-1 counter on target creature.",
									ManaCost: opt.Val(cost.Mana{
										cost.B,
									}),
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:        cost.AdditionalTap,
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Targets: []game.TargetSpec{
													game.TargetSpec{
														MinTargets: 1,
														MaxTargets: 1,
														Constraint: "target creature",
														Allow:      game.TargetAllowPermanent,
														Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.AddCounter{
															Amount:      game.Fixed(1),
															Object:      game.TargetPermanentReference(0),
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
			Enchanted creature has "{B}, {T}: Put a -1/-1 counter on target creature."
		`,
		},
	}
}
