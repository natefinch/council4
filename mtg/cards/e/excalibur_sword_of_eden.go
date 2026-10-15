package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ExcaliburSwordOfEden is the card definition for Excalibur, Sword of Eden.
//
// Type: Legendary Artifact — Equipment
// Cost: {12}
//
// Oracle text:
//
//	This spell costs {X} less to cast, where X is the total mana value of historic permanents you control. (Artifacts, legendaries, and Sagas are historic.)
//	Equipped creature gets +10/+0 and has vigilance.
//	Equip legendary creature {2}
var ExcaliburSwordOfEden = newExcaliburSwordOfEden

func newExcaliburSwordOfEden() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Excalibur, Sword of Eden",
			ManaCost: opt.Val(cost.Mana{
				cost.O(12),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {X} less to cast, where X is the total mana value of historic permanents you control. (Artifacts, legendaries, and Sagas are historic.)",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								DynamicReduction: &game.DynamicAmount{
									Kind:        game.DynamicAmountTotalManaValueInGroup,
									Multiplier:  1,
									CounterKind: counter.PlusOnePlusOne,
									Group:       game.BattlefieldGroup(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypes: []types.Card{types.Artifact}}, game.Selection{Supertypes: []types.Super{types.Legendary}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Saga")}}}, Controller: game.ControllerYou}),
								},
							},
						},
					},
				},
				game.StaticAbility{
					Text: "Equipped creature gets +10/+0 and has vigilance.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 10,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Vigilance},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipRestrictedActivatedAbility(cost.Mana{cost.O(2)}, []types.Super{types.Legendary}, nil),
			},
			OracleText: `
			This spell costs {X} less to cast, where X is the total mana value of historic permanents you control. (Artifacts, legendaries, and Sagas are historic.)
			Equipped creature gets +10/+0 and has vigilance.
			Equip legendary creature {2}
		`,
		},
	}
}
