package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DomriRade is the card definition for Domri Rade.
//
// Type: Legendary Planeswalker — Domri
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	+1: Look at the top card of your library. If it's a creature card, you may reveal it and put it into your hand.
//	−2: Target creature you control fights another target creature.
//	−7: You get an emblem with "Creatures you control have double strike, trample, hexproof, and haste."
var DomriRade = newDomriRade

func newDomriRade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Domri Rade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Domri},
			Loyalty:    opt.Val(3),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+1: Look at the top card of your library. If it's a creature card, you may reveal it and put it into your hand.",
					LoyaltyCost: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LookAtLibraryTop{
											Player:        game.ControllerReference(),
											PublishLinked: game.LinkedKey("sequence-effect-0-product"),
										},
										LocalProducts: game.LocalProducts{
											Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
										},
									},
									game.Instruction{
										Primitive: game.Reveal{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "sequence-effect-0-product",
											},
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "If it's a creature card",
												Object:            opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
												ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
										PublishCondition:        game.ConditionKey("condition-0"),
										Optional:                true,
										PublishOptionalDecision: game.OptionalDecisionKey("optional-clause-2"),
									},
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "sequence-effect-0-product",
											},
											FromZone:    zone.Library,
											Destination: zone.Hand,
										},
										ConditionGate:        game.ConditionKey("condition-0"),
										OptionalDecisionGate: game.OptionalDecisionKey("optional-clause-2"),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−2: Target creature you control fights another target creature.",
					LoyaltyCost: -2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "Target creature you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									},
									game.TargetSpec{
										MinTargets:               1,
										MaxTargets:               1,
										Constraint:               "another target creature",
										Allow:                    game.TargetAllowPermanent,
										Selection:                opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
										DistinctFromPriorTargets: true,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Fight{
											Object:        game.TargetPermanentReference(0),
											RelatedObject: game.TargetPermanentReference(1),
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
					Text:        "−7: You get an emblem with \"Creatures you control have double strike, trample, hexproof, and haste.\"",
					LoyaltyCost: -7,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.StaticAbility{
													Text: "Creatures you control have double strike, trample, hexproof, and haste.",
													ContinuousEffects: []game.ContinuousEffect{
														game.ContinuousEffect{
															Duration:    game.DurationPermanent,
															Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
															Layer:       game.LayerAbility,
															AddKeywords: []game.Keyword{game.DoubleStrike, game.Trample, game.Hexproof, game.Haste},
														},
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
			+1: Look at the top card of your library. If it's a creature card, you may reveal it and put it into your hand.
			−2: Target creature you control fights another target creature.
			−7: You get an emblem with "Creatures you control have double strike, trample, hexproof, and haste."
		`,
		},
	}
}
