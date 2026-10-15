package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FriendlyNeighborhood is the card definition for Friendly Neighborhood.
//
// Type: Enchantment — Aura
// Cost: {3}{W}
//
// Oracle text:
//
//	Enchant land
//	When this Aura enters, create three 1/1 green and white Human Citizen creature tokens.
//	Enchanted land has "{1}, {T}: Target creature gets +1/+1 until end of turn for each creature you control. Activate only as a sorcery."
var FriendlyNeighborhood = newFriendlyNeighborhood

func newFriendlyNeighborhood() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Friendly Neighborhood",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "land",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
				}),
				game.StaticAbility{
					Text: "Enchanted land has \"{1}, {T}: Target creature gets +1/+1 until end of turn for each creature you control. Activate only as a sorcery.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ActivatedAbility{
									Text: "{1}, {T}: Target creature gets +1/+1 until end of turn for each creature you control. Activate only as a sorcery.",
									ManaCost: opt.Val(cost.Mana{
										cost.O(1),
									}),
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:        cost.AdditionalTap,
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Timing:         game.SorceryOnly,
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
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
													game.Instruction{
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this Aura enters, create three 1/1 green and white Human Citizen creature tokens.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(3),
											Source: game.TokenDef(friendlyNeighborhoodToken),
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
			Enchant land
			When this Aura enters, create three 1/1 green and white Human Citizen creature tokens.
			Enchanted land has "{1}, {T}: Target creature gets +1/+1 until end of turn for each creature you control. Activate only as a sorcery."
		`,
		},
	}
}

var friendlyNeighborhoodToken = newFriendlyNeighborhoodToken()

func newFriendlyNeighborhoodToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human Citizen",
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Citizen},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
