package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StingTheGlintingDagger is the card definition for Sting, the Glinting Dagger.
//
// Type: Legendary Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature gets +1/+1 and has haste.
//	At the beginning of each combat, untap equipped creature.
//	Equipped creature has first strike as long as it's blocking or blocked by a Goblin or Orc.
//	Equip {2}
var StingTheGlintingDagger = newStingTheGlintingDagger

func newStingTheGlintingDagger() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sting, the Glinting Dagger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+1 and has haste.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Haste},
						},
					},
				},
				game.StaticAbility{
					Text: "Equipped creature has first strike as long as it's blocking or blocked by a Goblin or Orc.",
					Condition: opt.Val(game.Condition{
						SourceCounterKind:                       counter.PlusOnePlusOne,
						SourceAttachedCombatCounterpartSubtypes: [2]types.Sub{0: types.Goblin, 1: types.Orc},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.FirstStrike},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of each combat, untap equipped creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepBeginningOfCombat,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Object: game.SourceAttachedPermanentReference(),
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
			Equipped creature gets +1/+1 and has haste.
			At the beginning of each combat, untap equipped creature.
			Equipped creature has first strike as long as it's blocking or blocked by a Goblin or Orc.
			Equip {2}
		`,
		},
	}
}
