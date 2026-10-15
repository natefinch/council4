package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrimalAdversary is the card definition for Primal Adversary.
//
// Type: Creature — Wolf
// Cost: {2}{G}
//
// Oracle text:
//
//	Trample
//	When this creature enters, you may pay {1}{G} any number of times. When you pay this cost one or more times, put that many +1/+1 counters on this creature, then up to that many target lands you control become 3/3 Wolf creatures with haste that are still lands.
var PrimalAdversary = newPrimalAdversary

func newPrimalAdversary() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Primal Adversary",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, you may pay {1}{G} any number of times. When you pay this cost one or more times, put that many +1/+1 counters on this creature, then up to that many target lands you control become 3/3 Wolf creatures with haste that are still lands.",
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
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PayRepeatedly{
											Payment: game.ResolutionPayment{
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
													cost.G,
												}),
											},
											PublishCount: game.ResultKey("pay-repeatedly-count"),
										},
									},
									game.Instruction{
										Primitive: game.AddCounter{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:      game.DynamicAmountChosenNumber,
												ResultKey: game.ResultKey("pay-repeatedly-count"),
											}),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerType,
													AddTypes:    []types.Card{types.Creature},
													AddSubtypes: []types.Sub{types.Wolf},
												},
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Haste},
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerPowerToughnessSet,
													SetPower: opt.Val(game.PT{
														Value: 3,
													}),
													SetToughness: opt.Val(game.PT{
														Value: 3,
													}),
												},
											},
											Duration:   game.DurationPermanent,
											ChooseFrom: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Land}}),
											ChooseUpTo: game.Dynamic(game.DynamicAmount{
												Kind:      game.DynamicAmountChosenNumber,
												ResultKey: game.ResultKey("pay-repeatedly-count"),
											}),
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
			Trample
			When this creature enters, you may pay {1}{G} any number of times. When you pay this cost one or more times, put that many +1/+1 counters on this creature, then up to that many target lands you control become 3/3 Wolf creatures with haste that are still lands.
		`,
		},
	}
}
