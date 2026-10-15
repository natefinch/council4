package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UlamogSCrusher is the card definition for Ulamog's Crusher.
//
// Type: Creature — Eldrazi
// Cost: {8}
//
// Oracle text:
//
//	Annihilator 2 (Whenever this creature attacks, defending player sacrifices two permanents of their choice.)
//	This creature attacks each combat if able.
var UlamogSCrusher = newUlamogSCrusher

func newUlamogSCrusher() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ulamog's Crusher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.MustAttackStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, defending player sacrifices two permanents of their choice.",
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
											Amount: game.Fixed(2),
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
			Annihilator 2 (Whenever this creature attacks, defending player sacrifices two permanents of their choice.)
			This creature attacks each combat if able.
		`,
		},
	}
}
