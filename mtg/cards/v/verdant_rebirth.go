package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VerdantRebirth is the card definition for Verdant Rebirth.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Until end of turn, target creature gains "When this creature dies, return it to its owner's hand."
//	Draw a card.
var VerdantRebirth = newVerdantRebirth

func newVerdantRebirth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Verdant Rebirth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
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
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									AddAbilities: []game.Ability{
										new(game.TriggeredAbility{
											Trigger: game.TriggerCondition{
												Type: game.TriggerWhen,
												Pattern: game.TriggerPattern{
													Event:            game.EventPermanentDied,
													Source:           game.TriggerSourceSelf,
													SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
												},
											},
											Content: game.Mode{
												Sequence: []game.Instruction{
													{
														Primitive: game.MoveCard{
															Card:        game.CardReference{Kind: game.CardReferenceEvent},
															FromZone:    zone.Graveyard,
															Destination: zone.Hand,
														},
													},
												},
											}.Ability(),
										}),
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Until end of turn, target creature gains "When this creature dies, return it to its owner's hand."
			Draw a card.
		`,
		},
	}
}
