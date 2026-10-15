package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoriaMarauder is the card definition for Moria Marauder.
//
// Type: Creature — Goblin Warrior
// Cost: {R}{R}
//
// Oracle text:
//
//	Double strike
//	Whenever a Goblin or Orc you control deals combat damage to a player, exile the top card of your library. You may play that card this turn.
var MoriaMarauder = newMoriaMarauder

func newMoriaMarauder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Moria Marauder",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a Goblin or Orc you control deals combat damage to a player, exile the top card of your library. You may play that card this turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							StackObjectKind:       game.StackSpell,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{SubtypesAny: []types.Sub{types.Goblin, types.Orc}},
							RequireCombatDamage:   true,
							CounterKind:           counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ImpulseExile{
											Player:   game.ControllerReference(),
											Amount:   game.Fixed(1),
											Duration: game.DurationThisTurn,
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
			Double strike
			Whenever a Goblin or Orc you control deals combat damage to a player, exile the top card of your library. You may play that card this turn.
		`,
		},
	}
}
