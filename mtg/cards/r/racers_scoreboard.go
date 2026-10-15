package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RacersScoreboard is the card definition for Racers' Scoreboard.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)
//	When this artifact enters, draw two cards, then discard a card.
//	Max speed — Spells you cast cost {1} less to cast.
var RacersScoreboard = newRacersScoreboard

func newRacersScoreboard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Racers' Scoreboard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Max speed — Spells you cast cost {1} less to cast.",
					Condition: opt.Val(game.Condition{
						SourceCounterKind:     counter.PlusOnePlusOne,
						ControllerHasMaxSpeed: true,
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.StartEnginesTriggeredBody,
				game.TriggeredAbility{
					Text: "When this artifact enters, draw two cards, then discard a card.",
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
										Primitive: game.Draw{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(1),
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
			Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)
			When this artifact enters, draw two cards, then discard a card.
			Max speed — Spells you cast cost {1} less to cast.
		`,
		},
	}
}
