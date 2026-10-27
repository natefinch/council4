// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTNumericCostSubject is the card definition for RT Numeric Cost Subject.
//
// Type: Artifact
//
// Oracle text:
//
//	{T}, Sacrifice a creature: Create a Food token. If the sacrificed creature had toughness 4 or greater, create two Food tokens instead.
var RTNumericCostSubject = newRTNumericCostSubject

func newRTNumericCostSubject() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Numeric Cost Subject",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice a creature: Create a Food token. If the sacrificed creature had toughness 4 or greater, create two Food tokens instead.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice a creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
							SubjectKey:         "paid-cost-1-2",
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(rTNumericCostSubjectToken),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:   "If the sacrificed creature had toughness 4 or greater",
												Negate: true,
												Object: opt.Val(game.PaidCostReference("paid-cost-1-2", game.PaidCostSacrifice, types.Creature)),
												ObjectMatches: opt.Val(game.Selection{Toughness: opt.Val(compare.Int{
													Op:    compare.GreaterOrEqual,
													Value: 4,
												})}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
										PublishCondition: game.ConditionKey("condition-0"),
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(2),
											Source: game.TokenDef(rTNumericCostSubjectToken),
										},
										ConditionGate:       game.ConditionKey("condition-0"),
										ConditionGateNegate: true,
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
			{T}, Sacrifice a creature: Create a Food token. If the sacrificed creature had toughness 4 or greater, create two Food tokens instead.
		`,
		},
	}
}

var rTNumericCostSubjectToken = newRTNumericCostSubjectToken()

func newRTNumericCostSubjectToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Food",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Food},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(3),
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
		},
	}
}
