// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTSacrificeCostSubject is the card definition for RT Sacrifice Cost Subject.
//
// Type: Artifact
//
// Oracle text:
//
//	Sacrifice a creature: You gain 1 life. If the sacrificed creature was red, draw a card.
var RTSacrificeCostSubject = newRTSacrificeCostSubject

func newRTSacrificeCostSubject() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Sacrifice Cost Subject",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice a creature: You gain 1 life. If the sacrificed creature was red, draw a card.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice a creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
							SubjectKey:         "paid-cost-1-1",
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "If the sacrificed creature was red",
												Object:            opt.Val(game.PaidCostReference("paid-cost-1-1", game.PaidCostSacrifice, types.Creature)),
												ObjectMatches:     opt.Val(game.Selection{ColorsAny: []color.Color{color.Red}}),
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
			Sacrifice a creature: You gain 1 life. If the sacrificed creature was red, draw a card.
		`,
		},
	}
}
