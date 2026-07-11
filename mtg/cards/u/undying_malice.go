package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UndyingMalice is the card definition for Undying Malice.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Until end of turn, target creature gains "When this creature dies, return it to the battlefield tapped under its owner's control with a +1/+1 counter on it."
var UndyingMalice = newUndyingMalice

func newUndyingMalice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Undying Malice",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
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
														Primitive: game.PutOnBattlefield{
															Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
															EntryTapped:   true,
															EntryCounters: []game.CounterPlacement{game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}},
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
				},
			}.Ability()),
			OracleText: `
			Until end of turn, target creature gains "When this creature dies, return it to the battlefield tapped under its owner's control with a +1/+1 counter on it."
		`,
		},
	}
}
