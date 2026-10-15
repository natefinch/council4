package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TombOfHorrorsAdventurer is the card definition for Tomb of Horrors Adventurer.
//
// Type: Creature — Elf Monk
// Cost: {5}{U}
//
// Oracle text:
//
//	When this creature enters, you take the initiative.
//	Whenever you cast your second spell each turn, copy it. If you've completed a dungeon, copy that spell twice instead. You may choose new targets for the copies. (A copy of a permanent spell becomes a token.)
var TombOfHorrorsAdventurer = newTombOfHorrorsAdventurer

func newTombOfHorrorsAdventurer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tomb of Horrors Adventurer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Monk},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, you take the initiative.",
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
										Primitive: game.TakeInitiative{
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
				game.TriggeredAbility{
					Text: "Whenever you cast your second spell each turn, copy it. If you've completed a dungeon, copy that spell twice instead. You may choose new targets for the copies. (A copy of a permanent spell becomes a token.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventSpellCast,
							Controller:                 game.TriggerControllerYou,
							StackObjectKind:            game.StackSpell,
							PlayerEventOrdinalThisTurn: 2,
							CounterKind:                counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CopyStackObject{
											Object:              game.EventStackObjectReference(),
											Count:               1,
											MayChooseNewTargets: true,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "If you've completed a dungeon",
												Negate:                      true,
												SourceCounterKind:           counter.PlusOnePlusOne,
												ControllerCompletedADungeon: true,
											}),
										}),
										PublishCondition: game.ConditionKey("condition-0"),
									},
									game.Instruction{
										Primitive: game.CopyStackObject{
											Object:              game.EventStackObjectReference(),
											Count:               2,
											MayChooseNewTargets: true,
										},
										ConditionGate:       game.ConditionKey("condition-0"),
										ConditionGateNegate: true,
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
			When this creature enters, you take the initiative.
			Whenever you cast your second spell each turn, copy it. If you've completed a dungeon, copy that spell twice instead. You may choose new targets for the copies. (A copy of a permanent spell becomes a token.)
		`,
		},
	}
}
