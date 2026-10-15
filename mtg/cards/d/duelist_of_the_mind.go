package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DuelistOfTheMind is the card definition for Duelist of the Mind.
//
// Type: Creature — Human Advisor
// Cost: {1}{U}
//
// Oracle text:
//
//	Flying, vigilance
//	Duelist of the Mind's power is equal to the number of cards you've drawn this turn.
//	Whenever you commit a crime, you may draw a card. If you do, discard a card. This ability triggers only once each turn.
var DuelistOfTheMind = newDuelistOfTheMind

func newDuelistOfTheMind() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Duelist of the Mind",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:       []color.Color{color.Blue},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Human, types.Advisor},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 3}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCardsDrawnThisTurn}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you commit a crime, you may draw a card. If you do, discard a card. This ability triggers only once each turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventCrimeCommitted,
							Player:          game.TriggerPlayerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										Optional:      true,
										PublishResult: game.ResultKey("if-you-do"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("if-you-do")},
										},
									},
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("if-you-do"),
											Succeeded: game.TriTrue,
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
			Flying, vigilance
			Duelist of the Mind's power is equal to the number of cards you've drawn this turn.
			Whenever you commit a crime, you may draw a card. If you do, discard a card. This ability triggers only once each turn.
		`,
		},
	}
}
