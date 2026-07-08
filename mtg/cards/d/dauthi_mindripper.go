package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DauthiMindripper is the card definition for Dauthi Mindripper.
//
// Type: Creature — Dauthi Minion
// Cost: {3}{B}
//
// Oracle text:
//
//	Shadow (This creature can block or be blocked by only creatures with shadow.)
//	Whenever this creature attacks and isn't blocked, you may sacrifice it. If you do, defending player discards three cards.
var DauthiMindripper = newDauthiMindripper

func newDauthiMindripper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dauthi Mindripper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dauthi, types.Minion},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ShadowStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerBecameUnblocked,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Sacrifice{
									Object: game.EventPermanentReference(),
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(3),
									Player: game.DefendingPlayerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Shadow (This creature can block or be blocked by only creatures with shadow.)
			Whenever this creature attacks and isn't blocked, you may sacrifice it. If you do, defending player discards three cards.
		`,
		},
	}
}
