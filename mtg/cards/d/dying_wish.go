package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DyingWish is the card definition for Dying Wish.
//
// Type: Enchantment — Aura
// Cost: {1}{B}
//
// Oracle text:
//
//	Enchant creature you control
//	When enchanted creature dies, target player loses X life and you gain X life, where X is its power.
var DyingWish = newDyingWish

func newDyingWish() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dying Wish",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
				}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When enchanted creature dies, target player loses X life and you gain X life, where X is its power.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceAttachedPermanent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target player",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountObjectPower,
												Multiplier: 1,
												Object:     game.EventPermanentReference(),
											}),
											Player: game.TargetPlayerReference(0),
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountObjectPower,
												Multiplier: 1,
												Object:     game.EventPermanentReference(),
											}),
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
			},
			OracleText: `
			Enchant creature you control
			When enchanted creature dies, target player loses X life and you gain X life, where X is its power.
		`,
		},
	}
}
