package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TragicSlip is the card definition for Tragic Slip.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Target creature gets -1/-1 until end of turn.
//	Morbid — That creature gets -13/-13 until end of turn instead if a creature died this turn.
var TragicSlip = newTragicSlip

func newTragicSlip() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Tragic Slip",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
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
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(-1),
							ToughnessDelta: game.Fixed(-1),
							Duration:       game.DurationUntilEndOfTurn,
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
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(-13),
							ToughnessDelta: game.Fixed(-13),
							Duration:       game.DurationUntilEndOfTurn,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets -1/-1 until end of turn.
			Morbid — That creature gets -13/-13 until end of turn instead if a creature died this turn.
		`,
		},
	}
}
