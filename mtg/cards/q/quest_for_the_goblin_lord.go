package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuestForTheGoblinLord is the card definition for Quest for the Goblin Lord.
//
// Type: Enchantment
// Cost: {R}
//
// Oracle text:
//
//	Whenever a Goblin you control enters, you may put a quest counter on this enchantment.
//	As long as this enchantment has five or more quest counters on it, creatures you control get +2/+0.
var QuestForTheGoblinLord = newQuestForTheGoblinLord

func newQuestForTheGoblinLord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Quest for the Goblin Lord",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as this enchantment has five or more quest counters on it, creatures you control get +2/+0.",
					Condition: opt.Val(game.Condition{
						Text:   "As long as this enchantment has five or more quest counters on it",
						Object: opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredCounter: counter.Quest, RequiredCounterCount: opt.Val(compare.Int{
							Op:    compare.GreaterOrEqual,
							Value: 5,
						})}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 2,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a Goblin you control enters, you may put a quest counter on this enchantment.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Goblin}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.Quest,
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
			Whenever a Goblin you control enters, you may put a quest counter on this enchantment.
			As long as this enchantment has five or more quest counters on it, creatures you control get +2/+0.
		`,
		},
	}
}
