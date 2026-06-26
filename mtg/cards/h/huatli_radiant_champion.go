package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HuatliRadiantChampion is the card definition for Huatli, Radiant Champion.
//
// Type: Legendary Planeswalker — Huatli
// Cost: {2}{G}{W}
//
// Oracle text:
//
//	+1: Put a loyalty counter on Huatli for each creature you control.
//	−1: Target creature gets +X/+X until end of turn, where X is the number of creatures you control.
//	−8: You get an emblem with "Whenever a creature you control enters, you may draw a card."
var HuatliRadiantChampion = newHuatliRadiantChampion()

func newHuatliRadiantChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Huatli, Radiant Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Huatli},
			Loyalty:    opt.Val(3),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Huatli")}, Controller: game.ControllerYou}),
									CounterKind: counter.Loyalty,
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -1,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object: game.TargetPermanentReference(0),
									PowerDelta: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
									ToughnessDelta: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -8,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateEmblem{
									EmblemAbilities: []game.Ability{
										new(game.TriggeredAbility{
											Trigger: game.TriggerCondition{
												Type: game.TriggerWhenever,
												Pattern: game.TriggerPattern{
													Event:            game.EventPermanentEnteredBattlefield,
													Controller:       game.TriggerControllerYou,
													SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
												},
											},
											Optional: true,
											Content: game.Mode{
												Sequence: []game.Instruction{
													{
														Primitive: game.Draw{
															Amount: game.Fixed(1),
															Player: game.ControllerReference(),
														},
													},
												},
											}.Ability(),
										}),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			+1: Put a loyalty counter on Huatli for each creature you control.
			−1: Target creature gets +X/+X until end of turn, where X is the number of creatures you control.
			−8: You get an emblem with "Whenever a creature you control enters, you may draw a card."
		`,
		},
	}
}
