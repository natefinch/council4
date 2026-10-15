package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WizardClass is the card definition for Wizard Class.
//
// Type: Enchantment — Class
// Cost: {U}
//
// Oracle text:
//
//	(Gain the next level as a sorcery to add its ability.)
//	You have no maximum hand size.
//	{2}{U}: Level 2
//	When this Class becomes level 2, draw two cards.
//	{4}{U}: Level 3
//	Whenever you draw a card, put a +1/+1 counter on target creature you control.
var WizardClass = newWizardClass

func newWizardClass() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Wizard Class",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Class},
			StaticAbilities: []game.StaticAbility{
				game.NoMaximumHandSizeStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{U}: Level 2",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.U,
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
					Text: "{4}{U}: Level 3",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
						cost.U,
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
					Text: "When this Class becomes level 2, draw two cards.",
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
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(2),
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
					Text: "Whenever you draw a card, put a +1/+1 counter on target creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventCardDrawn,
							Player:          game.TriggerPlayerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningCondition: opt.Val(game.Condition{
							SourceClassLevelAtLeast: 3,
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
										Constraint: "target creature you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
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
			You have no maximum hand size.
			{2}{U}: Level 2
			When this Class becomes level 2, draw two cards.
			{4}{U}: Level 3
			Whenever you draw a card, put a +1/+1 counter on target creature you control.
		`,
		},
	}
}
