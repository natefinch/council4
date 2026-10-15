package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnifyingTheory is the card definition for Unifying Theory.
//
// Type: Enchantment
// Cost: {1}{U}
//
// Oracle text:
//
//	Whenever a player casts a spell, that player may pay {2}. If the player does, they draw a card.
var UnifyingTheory = newUnifyingTheory

func newUnifyingTheory() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Unifying Theory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a player casts a spell, that player may pay {2}. If the player does, they draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {2}?",
												Payer:  opt.Val(game.EventPlayerReference()),
												ManaCost: opt.Val(cost.Mana{
													cost.O(2),
												}),
											},
										},
										PublishResult: game.ResultKey("event-player-paid"),
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.EventPlayerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("event-player-paid"),
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
			Whenever a player casts a spell, that player may pay {2}. If the player does, they draw a card.
		`,
		},
	}
}
