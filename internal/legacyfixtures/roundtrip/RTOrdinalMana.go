// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTOrdinalMana is the card definition for RT Ordinal Mana.
//
// Type: Artifact
//
// Oracle text:
//
//	{2}: Target creature gains trample until end of turn. If this is the third time this ability has resolved this turn, you may add {R}{R}{R}{R}{R}{R}{R}{R}.
var RTOrdinalMana = newRTOrdinalMana

func newRTOrdinalMana() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Ordinal Mana",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}: Target creature gains trample until end of turn. If this is the third time this ability has resolved this turn, you may add {R}{R}{R}{R}{R}{R}{R}{R}.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					ZoneOfFunction:            zone.Battlefield,
					CountsResolutionsThisTurn: true,
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
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Trample},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(8),
											ManaColor: mana.R,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                                   "If this is the third time this ability has resolved this turn",
												SourceCounterKind:                      counter.PlusOnePlusOne,
												SourceAbilityResolutionOrdinalThisTurn: 3,
											}),
										}),
										Optional: true,
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
			{2}: Target creature gains trample until end of turn. If this is the third time this ability has resolved this turn, you may add {R}{R}{R}{R}{R}{R}{R}{R}.
		`,
		},
	}
}
