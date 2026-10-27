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

// RTCounterQuantity is the card definition for RT Counter Quantity.
//
// Type: Artifact
//
// Oracle text:
//
//	{1}: Remove all charge counters from this artifact. Draw that many cards.
var RTCounterQuantity = newRTCounterQuantity

func newRTCounterQuantity() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Counter Quantity",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}: Remove all charge counters from this artifact. Draw that many cards.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.RemoveCounter{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:        game.DynamicAmountObjectCounters,
												CounterKind: counter.Charge,
												Object:      game.SourcePermanentReference(),
											}),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.Charge,
										},
										PublishResult: game.ResultKey("removed-counter-quantity-1"),
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountPreviousEffectResult,
												Multiplier: 1,
												ResultKey:  game.ResultKey("removed-counter-quantity-1"),
											}),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:             game.ResultKey("removed-counter-quantity-1"),
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
			{1}: Remove all charge counters from this artifact. Draw that many cards.
		`,
		},
	}
}
