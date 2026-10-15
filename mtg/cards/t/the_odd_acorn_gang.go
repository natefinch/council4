package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheOddAcornGang is the card definition for The Odd Acorn Gang.
//
// Type: Legendary Creature — Squirrel Warrior
// Cost: {3}{B}{G}
//
// Oracle text:
//
//	Reach, menace, trample
//	Squirrels you control have "{T}: Target Squirrel gets +2/+2 and gains trample until end of turn. Activate only as a sorcery."
//	Whenever one or more Squirrels you control deal combat damage to a player, draw a card.
var TheOddAcornGang = newTheOddAcornGang

func newTheOddAcornGang() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "The Odd Acorn Gang",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Squirrel, types.Warrior},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.MenaceStaticBody,
				game.TrampleStaticBody,
				game.StaticAbility{
					Text: "Squirrels you control have \"{T}: Target Squirrel gets +2/+2 and gains trample until end of turn. Activate only as a sorcery.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Squirrel")}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{T}: Target Squirrel gets +2/+2 and gains trample until end of turn. Activate only as a sorcery.",
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:        cost.AdditionalTap,
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Timing:         game.SorceryOnly,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Targets: []game.TargetSpec{
													game.TargetSpec{
														MinTargets: 1,
														MaxTargets: 1,
														Constraint: "target Squirrel",
														Allow:      game.TargetAllowPermanent,
														Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Squirrel}}),
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.ApplyContinuous{
															Object: opt.Val(game.TargetPermanentReference(0)),
															ContinuousEffects: []game.ContinuousEffect{
																game.ContinuousEffect{
																	Duration:       game.DurationPermanent,
																	Layer:          game.LayerPowerToughnessModify,
																	PowerDelta:     2,
																	ToughnessDelta: 2,
																},
																game.ContinuousEffect{
																	Duration:    game.DurationPermanent,
																	Layer:       game.LayerAbility,
																	AddKeywords: []game.Keyword{game.Trample},
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
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever one or more Squirrels you control deal combat damage to a player, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							StackObjectKind:       game.StackSpell,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{SubtypesAny: []types.Sub{types.Squirrel}},
							RequireCombatDamage:   true,
							OneOrMore:             true,
							CounterKind:           counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			Reach, menace, trample
			Squirrels you control have "{T}: Target Squirrel gets +2/+2 and gains trample until end of turn. Activate only as a sorcery."
			Whenever one or more Squirrels you control deal combat damage to a player, draw a card.
		`,
		},
	}
}
