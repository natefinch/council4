package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WickerWarcrawler is the card definition for Wicker Warcrawler.
//
// Type: Artifact Creature — Scarecrow
// Cost: {5}
//
// Oracle text:
//
//	Whenever this creature attacks or blocks, put a -1/-1 counter on it at end of combat.
var WickerWarcrawler = newWickerWarcrawler

func newWickerWarcrawler() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Wicker Warcrawler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks or blocks, put a -1/-1 counter on it at end of combat.",
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
																	Primitive: game.AddCounter{
																		Amount:      game.Fixed(1),
																		Object:      game.CapturedObjectReference(),
																		CounterKind: counter.MinusOneMinusOne,
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
			OracleText: `
			Whenever this creature attacks or blocks, put a -1/-1 counter on it at end of combat.
		`,
		},
	}
}
