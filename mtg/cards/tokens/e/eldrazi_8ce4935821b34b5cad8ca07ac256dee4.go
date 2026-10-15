package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Eldrazi is the card definition for Eldrazi.
//
// Type: Token Creature — Eldrazi
//
// Oracle text:
//   Annihilator 1 (Whenever this creature attacks, defending player sacrifices a permanent.)

// EldraziToken8ce4935821b34b5cad8ca07ac256dee4 is the card definition for Eldrazi.
var EldraziToken8ce4935821b34b5cad8ca07ac256dee4 = newEldraziToken8ce4935821b34b5cad8ca07ac256dee4()

func newEldraziToken8ce4935821b34b5cad8ca07ac256dee4() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, defending player sacrifices a permanent of their choice.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
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
										Primitive: game.SacrificePermanents{
											Player: game.DefendingPlayerReference(),
											Amount: game.Fixed(1),
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
			Annihilator 1 (Whenever this creature attacks, defending player sacrifices a permanent.)
		`,
		},
	}
}
