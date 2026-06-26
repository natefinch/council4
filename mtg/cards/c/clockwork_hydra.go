package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ClockworkHydra is the card definition for Clockwork Hydra.
//
// Type: Artifact Creature — Hydra
// Cost: {5}
//
// Oracle text:
//
//	This creature enters with four +1/+1 counters on it.
//	Whenever this creature attacks or blocks, remove a +1/+1 counter from it. If you do, it deals 1 damage to any target.
//	{T}: Put a +1/+1 counter on this creature.
var ClockworkHydra = newClockworkHydra()

func newClockworkHydra() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Clockwork Hydra",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Hydra},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Put a +1/+1 counter on this creature.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventBlockerDeclared,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.RemoveCounter{
									Amount:      game.Fixed(1),
									Object:      game.EventPermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(1),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.EventPermanentReference()),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with four +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 4}),
			},
			OracleText: `
			This creature enters with four +1/+1 counters on it.
			Whenever this creature attacks or blocks, remove a +1/+1 counter from it. If you do, it deals 1 damage to any target.
			{T}: Put a +1/+1 counter on this creature.
		`,
		},
	}
}
