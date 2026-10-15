package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AjaniWiseCounselor is the card definition for Ajani, Wise Counselor.
//
// Type: Legendary Planeswalker — Ajani
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	+2: You gain 1 life for each creature you control.
//	−3: Creatures you control get +2/+2 until end of turn.
//	−9: Put X +1/+1 counters on target creature, where X is your life total.
var AjaniWiseCounselor = newAjaniWiseCounselor

func newAjaniWiseCounselor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ajani, Wise Counselor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Ajani},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+2: You gain 1 life for each creature you control.",
					LoyaltyCost: 2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
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
				game.LoyaltyAbility{
					Text:        "−3: Creatures you control get +2/+2 until end of turn.",
					LoyaltyCost: -3,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:       game.DurationPermanent,
													Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
													Layer:          game.LayerPowerToughnessModify,
													PowerDelta:     2,
													ToughnessDelta: 2,
												},
											},
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
				game.LoyaltyAbility{
					Text:        "−9: Put X +1/+1 counters on target creature, where X is your life total.",
					LoyaltyCost: -9,
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
										Primitive: game.AddCounter{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountControllerLife,
												Multiplier: 1,
											}),
											Object:      game.TargetPermanentReference(0),
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
			},
			OracleText: `
			+2: You gain 1 life for each creature you control.
			−3: Creatures you control get +2/+2 until end of turn.
			−9: Put X +1/+1 counters on target creature, where X is your life total.
		`,
		},
	}
}
