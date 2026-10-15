package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AjaniAdversaryOfTyrants is the card definition for Ajani, Adversary of Tyrants.
//
// Type: Legendary Planeswalker — Ajani
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	+1: Put a +1/+1 counter on each of up to two target creatures.
//	−2: Return target creature card with mana value 2 or less from your graveyard to the battlefield.
//	−7: You get an emblem with "At the beginning of your end step, create three 1/1 white Cat creature tokens with lifelink."
var AjaniAdversaryOfTyrants = newAjaniAdversaryOfTyrants

func newAjaniAdversaryOfTyrants() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ajani, Adversary of Tyrants",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Ajani},
			Loyalty:    opt.Val(4),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+1: Put a +1/+1 counter on each of up to two target creatures.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 2,
										Constraint: "up to two target creatures",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
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
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(1),
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
				game.LoyaltyAbility{
					Text:        "−2: Return target creature card with mana value 2 or less from your graveyard to the battlefield.",
					LoyaltyCost: -2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature card with mana value 2 or less from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{
											Op:    compare.LessOrEqual,
											Value: 2,
										})}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−7: You get an emblem with \"At the beginning of your end step, create three 1/1 white Cat creature tokens with lifelink.\"",
					LoyaltyCost: -7,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.TriggeredAbility{
													Text: "At the beginning of your end step, create three 1/1 white Cat creature tokens with lifelink.",
													Trigger: game.TriggerCondition{
														Type: game.TriggerAt,
														Pattern: game.TriggerPattern{
															Event:           game.EventBeginningOfStep,
															Controller:      game.TriggerControllerYou,
															StackObjectKind: game.StackSpell,
															CounterKind:     counter.PlusOnePlusOne,
															Step:            game.StepEnd,
														},
													},
													Content: game.AbilityContent{
														Modes: []game.Mode{
															game.Mode{
																Sequence: []game.Instruction{
																	game.Instruction{
																		Primitive: game.CreateToken{
																			Amount: game.Fixed(3),
																			Source: game.TokenDef(ajaniAdversaryOfTyrantsToken),
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
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			+1: Put a +1/+1 counter on each of up to two target creatures.
			−2: Return target creature card with mana value 2 or less from your graveyard to the battlefield.
			−7: You get an emblem with "At the beginning of your end step, create three 1/1 white Cat creature tokens with lifelink."
		`,
		},
	}
}

var ajaniAdversaryOfTyrantsToken = newAjaniAdversaryOfTyrantsToken()

func newAjaniAdversaryOfTyrantsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cat",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
		},
	}
}
