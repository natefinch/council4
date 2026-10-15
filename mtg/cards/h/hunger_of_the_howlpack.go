package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HungerOfTheHowlpack is the card definition for Hunger of the Howlpack.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Put a +1/+1 counter on target creature.
//	Morbid — Put three +1/+1 counters on that creature instead if a creature died this turn.
var HungerOfTheHowlpack = newHungerOfTheHowlpack

func newHungerOfTheHowlpack() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hunger of the Howlpack",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(1),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate: true,
								EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
									Event:            game.EventPermanentDied,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
									StackObjectKind:  game.StackSpell,
									CounterKind:      counter.PlusOnePlusOne,
								}, Window: game.EventHistoryCurrentTurn}),
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(3),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Put a +1/+1 counter on target creature.
			Morbid — Put three +1/+1 counters on that creature instead if a creature died this turn.
		`,
		},
	}
}
