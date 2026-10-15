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

// DryadGreenseeker is the card definition for Dryad Greenseeker.
//
// Type: Creature — Dryad
// Cost: {1}{G}
//
// Oracle text:
//
//	{T}: Look at the top card of your library. If it's a land card, you may reveal it and put it into your hand.
var DryadGreenseeker = newDryadGreenseeker

func newDryadGreenseeker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Dryad Greenseeker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dryad},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Look at the top card of your library. If it's a land card, you may reveal it and put it into your hand.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
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
												Text:              "If it's a land card",
												Object:            opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
												ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
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
			},
			OracleText: `
			{T}: Look at the top card of your library. If it's a land card, you may reveal it and put it into your hand.
		`,
		},
	}
}
