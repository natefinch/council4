package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ClockworkCondor is the card definition for Clockwork Condor.
//
// Type: Artifact Creature — Bird
// Cost: {4}
//
// Oracle text:
//
//	Flying
//	This creature enters with three +1/+1 counters on it.
//	Whenever this creature attacks or blocks, remove a +1/+1 counter from it at end of combat.
var ClockworkCondor = newClockworkCondor

func newClockworkCondor() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Clockwork Condor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks or blocks, remove a +1/+1 counter from it at end of combat.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							UnionEvent:      game.EventBlockerDeclared,
							BlockerScope:    game.BlockerTriggerOnce,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateDelayedTrigger{
											Trigger: game.DelayedTriggerDef{
												Timing: game.DelayedAtEndOfCombat,
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.RemoveCounter{
																		Amount:      game.Fixed(1),
																		Object:      game.CapturedObjectReference(),
																		CounterKind: counter.PlusOnePlusOne,
																	},
																},
															},
														},
													},
													MinModes: 1,
													MaxModes: 1,
												},
												CapturedObject: opt.Val(game.SourcePermanentReference()),
											},
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
				game.EntersWithCountersReplacement("This creature enters with three +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
			},
			OracleText: `
			Flying
			This creature enters with three +1/+1 counters on it.
			Whenever this creature attacks or blocks, remove a +1/+1 counter from it at end of combat.
		`,
		},
	}
}
