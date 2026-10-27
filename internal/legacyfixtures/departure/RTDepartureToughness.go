// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTDepartureToughness is the card definition for RT Departure Toughness.
//
// Type: Creature
//
// Oracle text:
//
//	When this creature enters, return target creature card from your graveyard to your hand. You gain life equal to that card's toughness.
var RTDepartureToughness = newRTDepartureToughness

func newRTDepartureToughness() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "RT Departure Toughness",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, return target creature card from your graveyard to your hand. You gain life equal to that card's toughness.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
											PublishDepartureCharacteristics: game.LibraryCardCharacteristics{
												Toughness: game.ResultKey("sequence-effect-0-departure-toughness"),
											},
										},
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("sequence-effect-0-departure-toughness")},
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountPreviousEffectResult,
												Multiplier: 1,
												ResultKey:  game.ResultKey("sequence-effect-0-departure-toughness"),
											}),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:             game.ResultKey("sequence-effect-0-departure-toughness"),
											AmountAvailable: true,
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
			When this creature enters, return target creature card from your graveyard to your hand. You gain life equal to that card's toughness.
		`,
		},
	}
}
