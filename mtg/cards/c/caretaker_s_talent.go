package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CaretakerSTalent is the card definition for Caretaker's Talent.
//
// Type: Enchantment — Class
// Cost: {2}{W}
//
// Oracle text:
//
//	(Gain the next level as a sorcery to add its ability.)
//	Whenever one or more tokens you control enter, draw a card. This ability triggers only once each turn.
//	{W}: Level 2
//	When this Class becomes level 2, create a token that's a copy of target token you control.
//	{3}{W}: Level 3
//	Creature tokens you control get +2/+2.
var CaretakerSTalent = newCaretakerSTalent

func newCaretakerSTalent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Caretaker's Talent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Class},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creature tokens you control get +2/+2.",
					Condition: opt.Val(game.Condition{
						SourceClassLevelAtLeast: 3,
						SourceCounterKind:       counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{W}: Level 2",
					ManaCost: opt.Val(cost.Mana{
						cost.W,
					}),
					Timing: game.SorceryOnly,
					ActivationCondition: opt.Val(game.Condition{
						SourceClassLevelLessThan: 2,
						SourceCounterKind:        counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SetClassLevel{
											Object: game.SourcePermanentReference(),
											Amount: game.Fixed(2),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ActivatedAbility{
					Text: "{3}{W}: Level 3",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.W,
					}),
					Timing: game.SorceryOnly,
					ActivationCondition: opt.Val(game.Condition{
						SourceClassLevelAtLeast:  2,
						SourceClassLevelLessThan: 3,
						SourceCounterKind:        counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SetClassLevel{
											Object: game.SourcePermanentReference(),
											Amount: game.Fixed(3),
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever one or more tokens you control enter, draw a card. This ability triggers only once each turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{TokenOnly: true},
							StackObjectKind:  game.StackSpell,
							OneOrMore:        true,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					MaxTriggersPerTurn: 1,
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
				game.TriggeredAbility{
					Text: "When this Class becomes level 2, create a token that's a copy of target token you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventClassLevelGained,
							Source:           game.TriggerSourceSelf,
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
							ClassBecameLevel: 2,
						},
						InterveningCondition: opt.Val(game.Condition{
							SourceClassLevelAtLeast: 2,
							SourceCounterKind:       counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target token you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{TokenOnly: true, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenCopyOf(game.TokenCopySpec{
												Source: game.TokenCopySourceObject,
												Object: game.TargetPermanentReference(0),
											}),
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
			(Gain the next level as a sorcery to add its ability.)
			Whenever one or more tokens you control enter, draw a card. This ability triggers only once each turn.
			{W}: Level 2
			When this Class becomes level 2, create a token that's a copy of target token you control.
			{3}{W}: Level 3
			Creature tokens you control get +2/+2.
		`,
		},
	}
}
