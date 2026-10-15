package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DualCasting is the card definition for Dual Casting.
//
// Type: Enchantment — Aura
// Cost: {1}{R}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature has "{R}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy."
var DualCasting = newDualCasting

func newDualCasting() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Dual Casting",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
					Text: "Enchanted creature has \"{R}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{R}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy.",
									ManaCost: opt.Val(cost.Mana{
										cost.R,
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
														Constraint: "target instant or sorcery spell you control",
														Allow:      game.TargetAllowStackObject,
														Predicate: game.TargetPredicate{
															SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
															StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
															Controller:        game.ControllerYou,
														},
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.CopyStackObject{
															Object:              game.TargetStackObjectReference(0),
															Count:               1,
															MayChooseNewTargets: true,
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
			Enchanted creature has "{R}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy."
		`,
		},
	}
}
