package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InnkeeperSTalent is the card definition for Innkeeper's Talent.
//
// Type: Enchantment — Class
// Cost: {1}{G}
//
// Oracle text:
//
//	(Gain the next level as a sorcery to add its ability.)
//	At the beginning of combat on your turn, put a +1/+1 counter on target creature you control.
//	{G}: Level 2
//	Permanents you control with counters on them have ward {1}.
//	{3}{G}: Level 3
//	If you would put one or more counters on a permanent or player, put twice that many of each of those kinds of counters on that permanent or player instead.
var InnkeeperSTalent = newInnkeeperSTalent

func newInnkeeperSTalent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Innkeeper's Talent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Class},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						SourceClassLevelAtLeast: 2,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{MatchAnyCounter: true}),
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(1)})),
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{G}: Level 2",
					ManaCost: opt.Val(cost.Mana{cost.G}),
					Timing:   game.SorceryOnly,
					ActivationCondition: opt.Val(game.Condition{
						SourceClassLevelLessThan: 2,
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SetClassLevel{
									Object: game.SourcePermanentReference(),
									Amount: game.Fixed(2),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:     "{3}{G}: Level 3",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.G}),
					Timing:   game.SorceryOnly,
					ActivationCondition: opt.Val(game.Condition{
						SourceClassLevelAtLeast:  2,
						SourceClassLevelLessThan: 3,
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SetClassLevel{
									Object: game.SourcePermanentReference(),
									Amount: game.Fixed(3),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepBeginningOfCombat,
						},
					},
					Content: game.Mode{
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
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.ClassLevelGatedReplacement(game.AnyCounterPlacementReplacement("If you would put one or more counters on a permanent or player, put twice that many of each of those kinds of counters on that permanent or player instead.", 2, 0, game.TriggerControllerYou), 3),
			},
			OracleText: `
			(Gain the next level as a sorcery to add its ability.)
			At the beginning of combat on your turn, put a +1/+1 counter on target creature you control.
			{G}: Level 2
			Permanents you control with counters on them have ward {1}.
			{3}{G}: Level 3
			If you would put one or more counters on a permanent or player, put twice that many of each of those kinds of counters on that permanent or player instead.
		`,
		},
	}
}
