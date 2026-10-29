package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CrowdedCrypt is the card definition for Crowded Crypt.
//
// Type: Artifact
// Cost: {2}{B}
//
// Oracle text:
//
//	{T}: Add {B}.
//	Whenever a creature you control dies, put a corpse counter on this artifact.
//	{4}{B}{B}, {T}, Sacrifice this artifact: Create a 2/2 black Zombie creature token with decayed for each corpse counter on this artifact. (A creature with decayed can't block. When it attacks, sacrifice it at end of combat.)
var CrowdedCrypt = newCrowdedCrypt

func newCrowdedCrypt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Crowded Crypt",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}{B}{B}, {T}, Sacrifice this artifact: Create a 2/2 black Zombie creature token with decayed for each corpse counter on this artifact. (A creature with decayed can't block. When it attacks, sacrifice it at end of combat.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
						cost.B,
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this artifact",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:        game.DynamicAmountObjectCounters,
												Multiplier:  1,
												CounterKind: counter.Corpse,
												Object:      game.SourcePermanentReference(),
											}),
											Source: game.TokenDef(crowdedCryptToken),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a creature you control dies, put a corpse counter on this artifact.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
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
											CounterKind: counter.Corpse,
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
			{T}: Add {B}.
			Whenever a creature you control dies, put a corpse counter on this artifact.
			{4}{B}{B}, {T}, Sacrifice this artifact: Create a 2/2 black Zombie creature token with decayed for each corpse counter on this artifact. (A creature with decayed can't block. When it attacks, sacrifice it at end of combat.)
		`,
		},
	}
}

var crowdedCryptToken = newCrowdedCryptToken()

func newCrowdedCryptToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't block.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateDelayedTrigger{
											Trigger: game.DelayedTriggerDef{
												Timing: game.DelayedAtEndOfCombat,
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.Sacrifice{
																		Object: game.CapturedObjectReference(),
																	},
																},
															},
														},
													},
													MinModes: 1,
													MaxModes: 1,
												},
												CapturedObject: opt.Val(game.SourcePermanentReference()),
											},
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
	}
}
