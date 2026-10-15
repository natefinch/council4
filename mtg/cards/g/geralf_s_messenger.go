package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GeralfSMessenger is the card definition for Geralf's Messenger.
//
// Type: Creature — Zombie
// Cost: {B}{B}{B}
//
// Oracle text:
//
//	This creature enters tapped.
//	When this creature enters, target opponent loses 2 life.
//	Undying (When this creature dies, if it had no +1/+1 counters on it, return it to the battlefield under its owner's control with a +1/+1 counter on it.)
var GeralfSMessenger = newGeralfSMessenger

func newGeralfSMessenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Geralf's Messenger",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, target opponent loses 2 life.",
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
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target opponent",
										Allow:      game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Fixed(2),
											Player: game.TargetPlayerReference(0),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.UndyingTriggeredBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This creature enters tapped."),
			},
			OracleText: `
			This creature enters tapped.
			When this creature enters, target opponent loses 2 life.
			Undying (When this creature dies, if it had no +1/+1 counters on it, return it to the battlefield under its owner's control with a +1/+1 counter on it.)
		`,
		},
	}
}
