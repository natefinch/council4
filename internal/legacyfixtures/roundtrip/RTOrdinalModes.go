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

// RTOrdinalModes is the card definition for RT Ordinal Modes.
//
// Type: Artifact
//
// Oracle text:
//
//	{1}: Choose one —
//	• If this is the second time this ability has resolved this turn, draw a card, then draw a card.
//	• You gain 2 life.
var RTOrdinalModes = newRTOrdinalModes

func newRTOrdinalModes() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Ordinal Modes",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}: Choose one —\n• If this is the second time this ability has resolved this turn, draw a card, then draw a card.\n• You gain 2 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					ZoneOfFunction:            zone.Battlefield,
					CountsResolutionsThisTurn: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "If this is the second time this ability has resolved this turn, draw a card, then draw a card.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                                   "If this is the second time this ability has resolved this turn",
												SourceCounterKind:                      counter.PlusOnePlusOne,
												SourceAbilityResolutionOrdinalThisTurn: 2,
											}),
										}),
										PublishCondition: game.ConditionKey("condition-0"),
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										ConditionGate: game.ConditionKey("condition-0"),
									},
								},
							},
							game.Mode{
								Text: "You gain 2 life.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
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
			{1}: Choose one —
			• If this is the second time this ability has resolved this turn, draw a card, then draw a card.
			• You gain 2 life.
		`,
		},
	}
}
