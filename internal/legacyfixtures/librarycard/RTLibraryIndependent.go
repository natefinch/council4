// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTLibraryIndependent is the card definition for RT Library Independent.
//
// Type: Artifact
//
// Oracle text:
//
//	{T}: Look at the top card of your library. If it's a land card, put it into your hand. Reveal the top card of your library. Scry 1. If it's a Zombie card, you gain 2 life.
var RTLibraryIndependent = newRTLibraryIndependent

func newRTLibraryIndependent() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Library Independent",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Look at the top card of your library. If it's a land card, put it into your hand. Reveal the top card of your library. Scry 1. If it's a Zombie card, you gain 2 life.",
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
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "sequence-effect-0-product",
											},
											FromZone:    zone.Library,
											Destination: zone.Hand,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "If it's a land card",
												Object:            opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
												ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.Reveal{
											Amount:        game.Fixed(1),
											Player:        game.ControllerReference(),
											PublishLinked: game.LinkedKey("sequence-effect-2-product"),
										},
										LocalProducts: game.LocalProducts{
											Links: []game.LinkedKey{game.LinkedKey("sequence-effect-2-product")},
										},
									},
									game.Instruction{
										Primitive: game.Scry{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "If it's a Zombie card",
												Object:            opt.Val(game.LinkedObjectReference("sequence-effect-2-product")),
												ObjectMatches:     opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Zombie}}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
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
			{T}: Look at the top card of your library. If it's a land card, put it into your hand. Reveal the top card of your library. Scry 1. If it's a Zombie card, you gain 2 life.
		`,
		},
	}
}
