package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FanaticOfXenagos is the card definition for Fanatic of Xenagos.
//
// Type: Creature — Centaur Warrior
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	Trample
//	Tribute 1 (As this creature enters, an opponent of your choice may put a +1/+1 counter on it.)
//	When this creature enters, if tribute wasn't paid, it gets +1/+1 and gains haste until end of turn.
var FanaticOfXenagos = newFanaticOfXenagos

func newFanaticOfXenagos() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Fanatic of Xenagos",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Centaur, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, if tribute wasn't paid, it gets +1/+1 and gains haste until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if tribute wasn't paid",
						InterveningCondition: opt.Val(game.Condition{
							Text:                 "if tribute wasn't paid",
							SourceCounterKind:    counter.PlusOnePlusOne,
							SourceTributeNotPaid: true,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.EventPermanentReference(),
											PowerDelta:     game.Fixed(1),
											ToughnessDelta: game.Fixed(1),
											Duration:       game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.EventPermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Haste},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.TributeReplacement("As this creature enters, an opponent of your choice may put 1 +1/+1 counters on it.", 1),
			},
			OracleText: `
			Trample
			Tribute 1 (As this creature enters, an opponent of your choice may put a +1/+1 counter on it.)
			When this creature enters, if tribute wasn't paid, it gets +1/+1 and gains haste until end of turn.
		`,
		},
	}
}
