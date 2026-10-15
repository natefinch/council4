package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OldGrowthEducator is the card definition for Old-Growth Educator.
//
// Type: Creature — Treefolk Druid
// Cost: {2}{B}{G}
//
// Oracle text:
//
//	Reach, vigilance
//	Infusion — When this creature enters, put two +1/+1 counters on it if you gained life this turn.
var OldGrowthEducator = newOldGrowthEducator

func newOldGrowthEducator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Old-Growth Educator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk, types.Druid},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Infusion — When this creature enters, put two +1/+1 counters on it if you gained life this turn.",
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
										Primitive: game.AddCounter{
											Amount:      game.Fixed(2),
											Object:      game.EventPermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "if you gained life this turn",
												SourceCounterKind: counter.PlusOnePlusOne,
												EventHistory: opt.Val(game.EventHistoryCondition{
													Pattern: game.TriggerPattern{
														Event:           game.EventLifeGained,
														Player:          game.TriggerPlayerYou,
														StackObjectKind: game.StackSpell,
														CounterKind:     counter.PlusOnePlusOne,
													},
													Window: game.EventHistoryCurrentTurn,
												}),
											}),
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
			Reach, vigilance
			Infusion — When this creature enters, put two +1/+1 counters on it if you gained life this turn.
		`,
		},
	}
}
