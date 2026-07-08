package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrownOfGondor is the card definition for Crown of Gondor.
//
// Type: Legendary Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature gets +1/+1 for each creature you control.
//	When a legendary creature you control enters, if there is no monarch, you become the monarch.
//	Equip {4}. This ability costs {3} less to activate if you're the monarch.
var CrownOfGondor = newCrownOfGondor

func newCrownOfGondor() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Crown of Gondor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerPowerToughnessModify,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipCostReductionActivatedAbility(cost.Mana{cost.O(4)}, game.CostModifier{
					Kind:             game.CostModifierAbility,
					GenericReduction: 3,
					ReductionCondition: opt.Val(game.Condition{
						ControllerIsMonarch: true,
					}),
				}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}},
						},
						InterveningIf: "if there is no monarch",
						InterveningCondition: opt.Val(game.Condition{
							NoMonarch: true,
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Equipped creature gets +1/+1 for each creature you control.
			When a legendary creature you control enters, if there is no monarch, you become the monarch.
			Equip {4}. This ability costs {3} less to activate if you're the monarch.
		`,
		},
	}
}
