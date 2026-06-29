package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HydraSGrowth is the card definition for Hydra's Growth.
//
// Type: Enchantment — Aura
// Cost: {2}{G}
//
// Oracle text:
//
//	Enchant creature
//	When this Aura enters, put a +1/+1 counter on enchanted creature.
//	At the beginning of your upkeep, double the number of +1/+1 counters on enchanted creature.
var HydraSGrowth = newHydraSGrowth()

func newHydraSGrowth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hydra's Growth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
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
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourceAttachedPermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									CounterKind: counter.PlusOnePlusOne,
									DoubleKind:  true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature
			When this Aura enters, put a +1/+1 counter on enchanted creature.
			At the beginning of your upkeep, double the number of +1/+1 counters on enchanted creature.
		`,
		},
	}
}
