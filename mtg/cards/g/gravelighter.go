package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Gravelighter is the card definition for Gravelighter.
//
// Type: Creature — Spirit
// Cost: {2}{B}
//
// Oracle text:
//
//	Flying
//	When this creature enters, draw a card if a creature died this turn. Otherwise, each player sacrifices a creature of their choice.
var Gravelighter = newGravelighter

func newGravelighter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gravelighter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, draw a card if a creature died this turn. Otherwise, each player sacrifices a creature of their choice.",
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
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "if a creature died this turn",
												SourceCounterKind: counter.PlusOnePlusOne,
												EventHistory: opt.Val(game.EventHistoryCondition{
													Pattern: game.TriggerPattern{
														Event:            game.EventPermanentDied,
														SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
														StackObjectKind:  game.StackSpell,
														CounterKind:      counter.PlusOnePlusOne,
													},
													Window: game.EventHistoryCurrentTurn,
												}),
											}),
										}),
										PublishCondition: game.ConditionKey("condition-0"),
									},
									game.Instruction{
										Primitive: game.SacrificePermanents{
											PlayerGroup: game.AllPlayersReference(),
											Amount:      game.Fixed(1),
											Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
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
			Flying
			When this creature enters, draw a card if a creature died this turn. Otherwise, each player sacrifices a creature of their choice.
		`,
		},
	}
}
