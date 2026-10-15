package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LivewireLash is the card definition for Livewire Lash.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature gets +2/+0 and has "Whenever this creature becomes the target of a spell, this creature deals 2 damage to any target."
//	Equip {2}
var LivewireLash = newLivewireLash

func newLivewireLash() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Livewire Lash",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +2/+0 and has \"Whenever this creature becomes the target of a spell, this creature deals 2 damage to any target.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 2,
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature becomes the target of a spell, this creature deals 2 damage to any target.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:                game.EventObjectBecameTarget,
											Source:               game.TriggerSourceSelf,
											MatchStackObjectKind: true,
											StackObjectKind:      game.StackSpell,
											CounterKind:          counter.PlusOnePlusOne,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Targets: []game.TargetSpec{
													game.TargetSpec{
														MinTargets: 1,
														MaxTargets: 1,
														Constraint: "any target",
														Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
														Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
													},
												},
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.Damage{
															Amount:       game.Fixed(2),
															Recipient:    game.AnyTargetDamageRecipient(0),
															DamageSource: opt.Val(game.SourcePermanentReference()),
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
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature gets +2/+0 and has "Whenever this creature becomes the target of a spell, this creature deals 2 damage to any target."
			Equip {2}
		`,
		},
	}
}
