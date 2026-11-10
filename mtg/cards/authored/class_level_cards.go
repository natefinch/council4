package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	buildersTalentOracleID = "0b19082c-8f80-448b-9e1b-0f324a159409"
	leadersTalentOracleID  = "134430b7-d566-485a-bd5c-7db8a285f4ca"
	paladinClassOracleID   = "ac9d52b2-78a5-4b17-8aaf-1c1cb936a508"
	gossipsTalentOracleID  = "96e6d220-87db-4dab-abaa-e80c9577d89a"
)

var BuildersTalent = cardset.Entry{
	OracleID:     buildersTalentOracleID,
	Name:         "Builder's Talent",
	Verification: "mtg/rules/class_level_cards_authored_test.go: TestBuildersTalentOwnText",
	New:          newBuildersTalent,
}

var LeadersTalent = cardset.Entry{
	OracleID:     leadersTalentOracleID,
	Name:         "Leader's Talent",
	Verification: "mtg/rules/class_level_cards_authored_test.go: TestLeadersTalentOwnText",
	New:          newLeadersTalent,
}

var PaladinClass = cardset.Entry{
	OracleID:     paladinClassOracleID,
	Name:         "Paladin Class",
	Verification: "mtg/rules/class_level_cards_authored_test.go: TestPaladinClassOwnText",
	New:          newPaladinClass,
}

var GossipsTalent = cardset.Entry{
	OracleID:     gossipsTalentOracleID,
	Name:         "Gossip's Talent",
	Verification: "mtg/rules/class_level_cards_authored_test.go: TestGossipsTalentOwnText",
	New:          newGossipsTalent,
}

func classLevelAbility(text string, manaCost cost.Mana, level int) game.ActivatedAbility {
	condition := game.Condition{SourceClassLevelLessThan: 2}
	if level == 3 {
		condition = game.Condition{
			SourceClassLevelAtLeast:  2,
			SourceClassLevelLessThan: 3,
		}
	}
	return game.ActivatedAbility{
		Text:                text,
		ManaCost:            opt.Val(manaCost),
		Timing:              game.SorceryOnly,
		ActivationCondition: opt.Val(condition),
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.SetClassLevel{
				Object: game.SourcePermanentReference(),
				Amount: game.Fixed(level),
			},
		}}}.Ability(),
	}
}

var buildersTalentWallToken = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Wall",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Wall")},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.DefenderStaticBody,
		},
	},
}

func newBuildersTalent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:     "Builder's Talent",
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Class},
			ActivatedAbilities: []game.ActivatedAbility{
				classLevelAbility("{W}: Level 2", cost.Mana{cost.W}, 2),
				classLevelAbility("{4}{W}: Level 3", cost.Mana{cost.O(4), cost.W}, 3),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: "When this Class enters, create a 0/4 white Wall creature token with defender.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(buildersTalentWallToken),
						},
					}}}.Ability(),
				},
				{
					Text: "Whenever one or more noncreature, nonland permanents you control enter, put a +1/+1 counter on target creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature, types.Land}},
							OneOrMore:        true,
						},
						InterveningCondition: opt.Val(game.Condition{SourceClassLevelAtLeast: 2}),
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target creature you control",
							Allow:      game.TargetAllowPermanent,
							Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.AddCounter{
								Amount:      game.Fixed(1),
								Object:      game.TargetPermanentReference(0),
								CounterKind: counter.PlusOnePlusOne,
							},
						}},
					}.Ability(),
				},
				{
					Text: "When this Class becomes level 3, return target noncreature, nonland permanent card from your graveyard to the battlefield.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventClassLevelGained,
							Source:           game.TriggerSourceSelf,
							ClassBecameLevel: 3,
						},
						InterveningCondition: opt.Val(game.Condition{SourceClassLevelAtLeast: 3}),
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target noncreature, nonland permanent card from your graveyard",
							Allow:      game.TargetAllowCard,
							TargetZone: zone.Graveyard,
							Selection: opt.Val(game.Selection{
								ExcludedTypes: []types.Card{types.Creature, types.Land},
								Controller:    game.ControllerYou,
							}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.PutOnBattlefield{
								Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							},
						}},
					}.Ability(),
				},
			},
			OracleText: "(Gain the next level as a sorcery to add its ability.)\n" +
				"When this Class enters, create a 0/4 white Wall creature token with defender.\n" +
				"{W}: Level 2\n" +
				"Whenever one or more noncreature, nonland permanents you control enter, put a +1/+1 counter on target creature you control.\n" +
				"{4}{W}: Level 3\n" +
				"When this Class becomes level 3, return target noncreature, nonland permanent card from your graveyard to the battlefield.",
		},
	}
}

func newLeadersTalent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:     "Leader's Talent",
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Class},
			ActivatedAbilities: []game.ActivatedAbility{
				classLevelAbility("{2}{W}: Level 2", cost.Mana{cost.O(2), cost.W}, 2),
				classLevelAbility("{3}{W}: Level 3", cost.Mana{cost.O(3), cost.W}, 3),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: "Whenever you attack, put a +1/+1 counter on target attacking creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Controller: game.TriggerControllerYou,
							OneOrMore:  true,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target attacking creature",
							Allow:      game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
								CombatState:   game.CombatStateAttacking,
							}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.AddCounter{
								Amount:      game.Fixed(1),
								Object:      game.TargetPermanentReference(0),
								CounterKind: counter.PlusOnePlusOne,
							},
						}},
					}.Ability(),
				},
				{
					Text: "Whenever a creature you control leaves the battlefield, if it had a counter on it, you gain 2 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							MatchFromZone:    true,
							FromZone:         zone.Battlefield,
						},
						InterveningCondition:                   opt.Val(game.Condition{SourceClassLevelAtLeast: 2}),
						InterveningIfEventPermanentHadCounters: true,
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					}}}.Ability(),
				},
				{
					Text: "Whenever you cast a spell, put a +1/+1 counter on each creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventSpellCast,
							Player: game.TriggerPlayerYou,
						},
						InterveningCondition: opt.Val(game.Condition{SourceClassLevelAtLeast: 3}),
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(1),
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							CounterKind: counter.PlusOnePlusOne,
						},
					}}}.Ability(),
				},
			},
			OracleText: "(Gain the next level as a sorcery to add its ability.)\n" +
				"Whenever you attack, put a +1/+1 counter on target attacking creature.\n" +
				"{2}{W}: Level 2\n" +
				"Whenever a creature you control leaves the battlefield, if it had a counter on it, you gain 2 life.\n" +
				"{3}{W}: Level 3\n" +
				"Whenever you cast a spell, put a +1/+1 counter on each creature you control.",
		},
	}
}

func newPaladinClass() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:     "Paladin Class",
			ManaCost: opt.Val(cost.Mana{cost.W}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Class},
			StaticAbilities: []game.StaticAbility{
				{
					Text: "Spells your opponents cast during your turn cost {1} more to cast.",
					RuleEffects: []game.RuleEffect{{
						Kind:                           game.RuleEffectCostModifier,
						AffectedPlayer:                 game.PlayerOpponent,
						RestrictedDuringControllerTurn: true,
						CostModifier: game.CostModifier{
							Kind:            game.CostModifierSpell,
							GenericIncrease: 1,
						},
					}},
				},
				{
					Text: "Creatures you control get +1/+1.",
					Condition: opt.Val(game.Condition{
						SourceClassLevelAtLeast: 2,
					}),
					ContinuousEffects: []game.ContinuousEffect{{
						Layer:          game.LayerPowerToughnessModify,
						Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						PowerDelta:     1,
						ToughnessDelta: 1,
					}},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				classLevelAbility("{2}{W}: Level 2", cost.Mana{cost.O(2), cost.W}, 2),
				classLevelAbility("{4}{W}: Level 3", cost.Mana{cost.O(4), cost.W}, 3),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: "Whenever you attack, until end of turn, target attacking creature gets +1/+1 for each other attacking creature and gains double strike.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Controller: game.TriggerControllerYou,
							OneOrMore:  true,
						},
						InterveningCondition: opt.Val(game.Condition{SourceClassLevelAtLeast: 3}),
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target attacking creature",
							Allow:      game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
								CombatState:   game.CombatStateAttacking,
							}),
						}},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object: game.TargetPermanentReference(0),
									PowerDelta: game.Dynamic(game.DynamicAmount{
										Kind:      game.DynamicAmountTriggeringAttackerCount,
										Selection: &game.Selection{RequiredTypes: []types.Card{types.Creature}},
										Addend:    -1,
									}),
									ToughnessDelta: game.Dynamic(game.DynamicAmount{
										Kind:      game.DynamicAmountTriggeringAttackerCount,
										Selection: &game.Selection{RequiredTypes: []types.Card{types.Creature}},
										Addend:    -1,
									}),
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{{
										Layer:       game.LayerAbility,
										AddKeywords: []game.Keyword{game.DoubleStrike},
									}},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: "(Gain the next level as a sorcery to add its ability.)\n" +
				"Spells your opponents cast during your turn cost {1} more to cast.\n" +
				"{2}{W}: Level 2\n" +
				"Creatures you control get +1/+1.\n" +
				"{4}{W}: Level 3\n" +
				"Whenever you attack, until end of turn, target attacking creature gets +1/+1 for each other attacking creature and gains double strike.",
		},
	}
}

const (
	gossipsTalentExileLink   game.LinkedKey = "gossips-talent/exiled"
	gossipsTalentExileResult game.ResultKey = "gossips-talent/exile-succeeded"
)

func newGossipsTalent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:     "Gossip's Talent",
			ManaCost: opt.Val(cost.Mana{cost.U}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Class},
			ActivatedAbilities: []game.ActivatedAbility{
				classLevelAbility("{1}{U}: Level 2", cost.Mana{cost.O(1), cost.U}, 2),
				classLevelAbility("{3}{U}: Level 3", cost.Mana{cost.O(3), cost.U}, 3),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: "Whenever a creature you control enters, surveil 1.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.Surveil{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					}}}.Ability(),
				},
				{
					Text: "Whenever you attack, target attacking creature with power 3 or less can't be blocked this turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Controller: game.TriggerControllerYou,
							OneOrMore:  true,
						},
						InterveningCondition: opt.Val(game.Condition{SourceClassLevelAtLeast: 2}),
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target attacking creature with power 3 or less",
							Allow:      game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
								CombatState:   game.CombatStateAttacking,
								Power:         opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3}),
							}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.ApplyRule{
								Object:      opt.Val(game.TargetPermanentReference(0)),
								RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectCantBeBlocked}},
								Duration:    game.DurationUntilEndOfTurn,
							},
						}},
					}.Ability(),
				},
				{
					Text: "Whenever a creature you control deals combat damage to a player, you may exile it, then return it to the battlefield under its owner's control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Subject:               game.TriggerSubjectDamageSource,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
							RequireCombatDamage:   true,
						},
						InterveningCondition: opt.Val(game.Condition{SourceClassLevelAtLeast: 3}),
					},
					Content: game.Mode{Sequence: []game.Instruction{
						{
							Primitive: game.MovePermanent{
								Object:        game.EventPermanentReference(),
								PublishLinked: gossipsTalentExileLink,
								Destination:   zone.Exile,
							},
							Optional:      true,
							PublishResult: gossipsTalentExileResult,
						},
						{
							Primitive: game.PutOnBattlefield{
								Source: game.LinkedBattlefieldSource(gossipsTalentExileLink),
							},
							ResultGate: opt.Val(game.InstructionResultGate{
								Key:       gossipsTalentExileResult,
								Succeeded: game.TriTrue,
							}),
						},
					}}.Ability(),
				},
			},
			OracleText: "(Gain the next level as a sorcery to add its ability.)\n" +
				"Whenever a creature you control enters, surveil 1.\n" +
				"{1}{U}: Level 2\n" +
				"Whenever you attack, target attacking creature with power 3 or less can't be blocked this turn.\n" +
				"{3}{U}: Level 3\n" +
				"Whenever a creature you control deals combat damage to a player, you may exile it, then return it to the battlefield under its owner's control.",
		},
	}
}
