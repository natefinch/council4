package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GylwainCastingDirector is the card definition for Gylwain, Casting Director.
//
// Type: Legendary Creature — Human Bard
// Cost: {1}{G}{W}
//
// Oracle text:
//
//	Whenever Gylwain or another nontoken creature you control enters, choose one —
//	• Create a Royal Role token attached to that creature.
//	• Create a Sorcerer Role token attached to that creature.
//	• Create a Monster Role token attached to that creature.
var GylwainCastingDirector = newGylwainCastingDirector

func newGylwainCastingDirector() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Gylwain, Casting Director",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Bard},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentEnteredBattlefield,
							Controller:             game.TriggerControllerYou,
							SubjectSelectionOrSelf: true,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Create a Royal Role token attached to that creature.",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount:          game.Fixed(1),
											Source:          game.TokenDef(gylwainCastingDirectorToken),
											EntryAttachedTo: opt.Val(game.EventPermanentReference()),
										},
									},
								},
							},
							game.Mode{
								Text: "Create a Sorcerer Role token attached to that creature.",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount:          game.Fixed(1),
											Source:          game.TokenDef(gylwainCastingDirectorToken2),
											EntryAttachedTo: opt.Val(game.EventPermanentReference()),
										},
									},
								},
							},
							game.Mode{
								Text: "Create a Monster Role token attached to that creature.",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount:          game.Fixed(1),
											Source:          game.TokenDef(gylwainCastingDirectorToken3),
											EntryAttachedTo: opt.Val(game.EventPermanentReference()),
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
			Whenever Gylwain or another nontoken creature you control enters, choose one —
			• Create a Royal Role token attached to that creature.
			• Create a Sorcerer Role token attached to that creature.
			• Create a Monster Role token attached to that creature.
		`,
		},
	}
}

var gylwainCastingDirectorToken = newGylwainCastingDirectorToken()

func newGylwainCastingDirectorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Royal Role",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Role},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(1)})),
							},
						},
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 and has ward {1}.
			(Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {1}.)
		`,
		},
	}
}

var gylwainCastingDirectorToken2 = newGylwainCastingDirectorToken2()

func newGylwainCastingDirectorToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Sorcerer Role",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Role},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:  game.EventAttackerDeclared,
											Source: game.TriggerSourceSelf,
										},
									},
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.Scry{
													Amount: game.Fixed(1),
													Player: game.ControllerReference(),
												},
											},
										},
									}.Ability(),
								}),
							},
						},
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 and has "Whenever this creature attacks, scry 1."
		`,
		},
	}
}

var gylwainCastingDirectorToken3 = newGylwainCastingDirectorToken3()

func newGylwainCastingDirectorToken3() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Monster Role",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Role},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Trample,
							},
						},
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 and has trample.
		`,
		},
	}
}
